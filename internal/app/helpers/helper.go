// Package helpers contain internal parts of general project logic
package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/consts"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
)

// encKey rand key
type encData struct {
	aesGCM cipher.AEAD
	nonce  []byte
}

// encInstance save encrypt data
var encInstance *encData

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// stringWithCharset generate rand string from charset
func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandomString Get rand string
func RandomString(length int) string {
	return stringWithCharset(length, charset)
}

// Decode userId  from encrypted cookie
func Decode(shaUserID string, userID *string) error {
	// Init encrypt data
	if err := keyInit(); err != nil {
		return err
	}
	// Convert to bytes from hex
	dst, err := hex.DecodeString(shaUserID)
	if err != nil {
		return err
	}
	// Decode
	src, err := encInstance.aesGCM.Open(nil, encInstance.nonce, dst, nil)
	if err != nil {
		return err
	}
	*userID = string(src)
	return nil
}

// Encode userId by GCM algorithm and get hex
func Encode(userID string) (string, error) {
	// Init encrypt data
	if err := keyInit(); err != nil {
		return "", err
	}
	src := []byte(userID)
	// Encrypt userId
	dst := encInstance.aesGCM.Seal(nil, encInstance.nonce, src, nil)
	// Get hexadecimal string from encode string
	sha := hex.EncodeToString(dst)
	return sha, nil
}

// generateRandom byte slice
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// keyInit init crypt params
func keyInit() error {
	// If you need generate new key
	if encInstance == nil {
		key, err := generateRandom(aes.BlockSize)
		if err != nil {
			return err
		}

		aesBlock, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		aesGCM, err := cipher.NewGCM(aesBlock)
		if err != nil {
			return err
		}
		// initialize vector
		nonce, err := generateRandom(aesGCM.NonceSize())
		if err != nil {
			return err
		}
		// Allocation enc type
		encInstance = new(encData)
		encInstance.aesGCM = aesGCM
		encInstance.nonce = nonce
	}
	return nil
}

// BodyFromJSON get bytes from JSON requests
func BodyFromJSON(w *http.ResponseWriter, r *http.Request) ([]byte, error) {
	var body []byte
	if r.Body == http.NoBody {
		http.Error(*w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return body, er.ErrBadResponse
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(*w, er.ErrUnknownURL.Error(), http.StatusBadRequest)
		return body, er.ErrUnknownURL
	}
	return body, nil
}

// GetContextUserID return uniq user id from session
func GetContextUserID(r *http.Request) user.UniqUser {
	userIDCtx := r.Context().Value(consts.UserIDCtxName)
	userID := "all"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}
	return user.UniqUser(userID)
}

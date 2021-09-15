package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"math/rand"
)

// encKey rand key
type encData struct {
	aesGCM cipher.AEAD
	nonce  []byte
}

// encInstance save encrypt data
var encInstance *encData

// Returns an int >= min, < max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// RandomString Generate a random string of A-Z chars with len = l
func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
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

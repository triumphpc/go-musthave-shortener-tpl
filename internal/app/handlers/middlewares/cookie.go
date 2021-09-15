// CookieMiddleware middleware set user encrypting uuid

package middlewares

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

// CookieUserIDName define cookie name for uuid
const CookieUserIDName = "UserId"

// ContextType set context name for user id
type ContextType string

var UserIDCtxName ContextType = "ctxUserId"

// CookieMiddleware check and set user token
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate new uuid
		userID := uuid.New().String()
		// Check if set cookie
		if cookieUserID, err := r.Cookie(CookieUserIDName); err == nil {
			userID = cookieUserID.Value
			//
			//logger.Info("cookieUserId", zap.String("cookieUserId", cookieUserID.Value))
			//_ = helpers.Decode(cookieUserID.Value, &userID)
		}
		// Generate hash from userId
		//encoded, err := helpers.Encode(userID)
		//logger.Info("User ID", zap.String("ID", userID))
		//logger.Info("User encoded", zap.String("Encoded", encoded))
		encoded := userID
		//if err == nil {
		cookie := &http.Cookie{
			Name:  CookieUserIDName,
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
		//} else {
		//	logger.Info("Encode cookie error", zap.Error(err))
		//}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDCtxName, userID)))
	})
}

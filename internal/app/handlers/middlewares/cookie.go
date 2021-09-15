// CookieMiddleware middleware set user encrypting uuid

package middlewares

import (
	"context"
	"github.com/google/uuid"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
)

// CookieUserIDName define cookie name for uuid
const CookieUserIDName = "UserId"

// UserIdCtxName set context name for user id
const UserIdCtxName = "ctxUserId"

// CookieMiddleware check and set user token
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate new uuid
		userId := uuid.New().String()
		// Check if set cookie
		if cookieUserId, err := r.Cookie(CookieUserIDName); err == nil {
			logger.Info("cookieUserId", zap.String("cookieUserId", cookieUserId.Value))
			_ = helpers.Decode(cookieUserId.Value, &userId)
		}
		// Generate hash from userId
		encoded, err := helpers.Encode(userId)
		logger.Info("User ID", zap.String("ID", userId))
		logger.Info("User encoded", zap.String("Encoded", encoded))
		if err == nil {
			cookie := &http.Cookie{
				Name:  CookieUserIDName,
				Value: encoded,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
		} else {
			logger.Info("Encode cookie error", zap.Error(err))
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIdCtxName, userId)))
	})
}

// CookieMiddleware middleware set user encrypting uuid

package middlewares

import (
	"context"
	"github.com/google/uuid"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"go.uber.org/zap"
	"net/http"
)

// CookieUserIDName define cookie name for uuid
const CookieUserIDName = "user_id"

// ContextType set context name for user id
type ContextType string

var UserIDCtxName ContextType = "ctxUserId"

type CookieMw struct {
	h http.Handler
	l *zap.Logger
}

func NewMw(l *zap.Logger) *CookieMw {
	return &CookieMw{l: l}
}

// CookieMiddleware check and set user token
func (h CookieMw) CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate new uuid
		userID := uuid.New().String()
		// Check if set cookie
		if cookieUserID, err := r.Cookie(CookieUserIDName); err == nil {
			h.l.Info("cookieUserId", zap.String("cookieUserId", cookieUserID.Value))
			_ = helpers.Decode(cookieUserID.Value, &userID)
		}
		// Generate hash from userId
		encoded, err := helpers.Encode(userID)
		h.l.Info("User ID", zap.String("ID", userID))
		h.l.Info("User encoded", zap.String("Encoded", encoded))
		if err == nil {
			cookie := &http.Cookie{
				Name:  CookieUserIDName,
				Value: encoded,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
		} else {
			h.l.Info("Encode cookie error", zap.Error(err))
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDCtxName, userID)))
	})
}

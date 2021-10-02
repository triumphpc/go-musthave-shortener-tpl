// CookieMiddleware middleware set user encrypting uuid

package middlewares

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/consts"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"go.uber.org/zap"
	"net/http"
)

type CookieMw struct {
	h http.Handler
	l *zap.Logger
}

func New(l *zap.Logger) *CookieMw {
	return &CookieMw{l: l}
}

// CookieMiddleware check and set user token
func (h CookieMw) CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate new uuid
		userID := uuid.New().String()
		// Check if set cookie
		if cookieUserID, err := r.Cookie(consts.CookieUserIDName); err == nil {
			h.l.Info("cookieUserId", zap.String("cookieUserId", cookieUserID.Value))
			_ = helpers.Decode(cookieUserID.Value, &userID)
		}
		// Generate hash from userId
		encoded, err := helpers.Encode(userID)
		h.l.Info("User ID", zap.String("ID", userID))
		h.l.Info("User encoded", zap.String("Encoded", encoded))
		if err == nil {
			cookie := &http.Cookie{
				Name:  consts.CookieUserIDName,
				Value: encoded,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
		} else {
			h.l.Info("Encode cookie error", zap.Error(err))
		}

		fmt.Println("COOKIE")
		fmt.Println(userID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), consts.UserIDCtxName, userID)))
	})
}

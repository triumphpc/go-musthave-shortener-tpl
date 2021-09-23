package ping

import (
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct {
	l *zap.Logger
}

func New(l *zap.Logger) *Handler {
	return &Handler{l}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := db.New(h.l)
	if err == nil {
		if err := conn.PingContext(r.Context()); err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	h.l.Info("not connect to db", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
}

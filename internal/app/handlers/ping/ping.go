package ping

import (
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct{}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := db.New()
	if err == nil {
		if err := conn.PingContext(r.Context()); err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	logger.Info("not connect to db", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
}

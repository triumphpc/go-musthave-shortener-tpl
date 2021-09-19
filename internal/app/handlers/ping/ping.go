package ping

import (
	"context"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct{}

func (h Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	conn, err := db.Instance()
	if err == nil {
		if err := conn.PingContext(context.Background()); err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	logger.Info("not connect to db", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
}

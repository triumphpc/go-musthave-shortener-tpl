package ping

import (
	"database/sql"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	db *sql.DB
	l  *zap.Logger
}

func New(db *sql.DB, l *zap.Logger) *Handler {
	return &Handler{db, l}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.db.PingContext(r.Context()); err == nil {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		h.l.Info("DB error", zap.Error(err))
	}
	w.WriteHeader(http.StatusInternalServerError)
}

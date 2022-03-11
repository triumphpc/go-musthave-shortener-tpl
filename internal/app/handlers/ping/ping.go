// Package ping implement handler for ping requests. Route /ping
package ping

import (
	"database/sql"
	"net/http"

	"go.uber.org/zap"
)

// Handler struct
type Handler struct {
	db *sql.DB
	l  *zap.Logger
}

// NewPing implement ping handler
func NewPing(db *sql.DB, l *zap.Logger) *Handler {
	return &Handler{db, l}
}

// ServeHTTP implement logic for ping hadler
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.db.PingContext(r.Context()); err == nil {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		h.l.Info("DB error", zap.Error(err))
	}
	w.WriteHeader(http.StatusInternalServerError)
}

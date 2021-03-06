// Package delete implement handler for delete links for route /api/user/urls
package delete

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
)

type Handler struct {
	l *zap.Logger
	p *worker.Pool
}

// New instance of deleted handler
func New(l *zap.Logger, p *worker.Pool) *Handler {
	return &Handler{l, p}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := helpers.BodyFromJSON(&w, r)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}

	var linkIDs []string
	err = json.Unmarshal(body, &linkIDs)
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	// Validate count
	if len(linkIDs) == 0 {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	// Add to pool ids on delete
	userID := helpers.GetContextUserID(r)
	if !h.p.Push(linkIDs, string(userID)) {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

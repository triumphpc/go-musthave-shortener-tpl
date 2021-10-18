package delete

import (
	"context"
	"encoding/json"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/mypool"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/repository"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct {
	l       *zap.Logger
	e       mypool.Executor
	storage repository.Repository
}

// New instance of deleted handler
func New(l *zap.Logger, e mypool.Executor, storage repository.Repository) *Handler {
	return &Handler{l, e, storage}
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

	err = h.e.Push(func(ctx context.Context) error {
		return h.storage.BunchUpdateAsDeleted(ctx, linkIDs, string(userID))
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

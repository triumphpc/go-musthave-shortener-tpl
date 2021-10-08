package delete

import (
	"database/sql"
	"encoding/json"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct {
	db *sql.DB
	l  *zap.Logger
	p  *worker.Pool
}

// New instance of deleted handler
func New(db *sql.DB, l *zap.Logger, p *worker.Pool) *Handler {
	return &Handler{db, l, p}
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
	if h.p.Push(linkIDs, string(userID)) == false {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

package delete

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

// workerCount count of worker for flow saving
const workerCount = 10

type Handler struct {
	db      *sql.DB
	l       *zap.Logger
	InputCh chan string
}

func New(db *sql.DB, l *zap.Logger) *Handler {
	return &Handler{db, l, nil}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := helpers.BodyFromJSON(&w, r)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}

	var correlationIDs []string
	err = json.Unmarshal(body, &correlationIDs)
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	// Validate count
	if len(correlationIDs) == 0 {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}

	h.InputCh = make(chan string)
	// Put in channel all ids
	go func() {
		for _, id := range correlationIDs {
			h.InputCh <- id
		}
		close(h.InputCh)
	}()
	// Distribution input ids from h.w.InputCh to 10 stacks in slice
	fanOutChs := h.fanOut()

	// fanOutChs range all slices
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	for _, fanOutCh := range fanOutChs {
		// To bunch saving
		h.fanInSave(ctx, fanOutCh, errCh, wg)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	if err := <-errCh; err != nil {
		h.l.Info("Handler error", zap.Error(err))
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		cancel()
		return
	}

	w.WriteHeader(http.StatusAccepted)
	cancel()
}

// FanInSave mark as delete for bunch
func (h Handler) fanInSave(ctx context.Context, input <-chan string, errCh chan<- error, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		var IDs []string
		var defErr error

		defer func() {
			if defErr != nil {
				select {
				case errCh <- defErr:
				case <-ctx.Done():
					h.l.Info("Aborting")
				}
			}
			wg.Done()
		}()

		for ID := range input {
			IDs = append(IDs, ID)
		}
		err := h.bunchUpdateAsDeleted(ctx, IDs)
		if err != nil {
			defErr = err
			return
		}
	}()
}

// bunchUpdateAsDeleted  update as deleted
func (h Handler) bunchUpdateAsDeleted(ctx context.Context, correlationIds []string) error {
	if len(correlationIds) == 0 {
		return nil
	}
	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	// Rollback handler
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)
	// Prepare statement
	query := "update storage.short_links set is_deleted=true where correlation_id = ANY($1)"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	// Close statement
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			h.l.Info("Close statement error", zap.Error(err))
		}
	}(stmt)

	// Update in transaction
	if _, err = stmt.ExecContext(ctx, pq.Array(correlationIds)); err != nil {
		return err
	}

	// Save changes
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// FanOut flow of ids
func (h Handler) fanOut() []chan string {
	// create stacks of chains
	cs := make([]chan string, 0, workerCount)
	for i := 0; i < workerCount; i++ {
		cs = append(cs, make(chan string))
	}
	// goroutines for channel stack distribution
	go func() {
		defer func(cs []chan string) {
			for _, c := range cs {
				close(c)
			}
		}(cs)

		for i := 0; i < len(cs); i++ {
			if i == len(cs)-1 {
				i = 0
			}
			// get id link from chan. If is not exist - out and close channels
			id, ok := <-h.InputCh
			if !ok {
				return
			}
			// distribution in stack
			cs[i] <- id
		}
	}()
	return cs
}

package routes

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/delete"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/ping"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"go.uber.org/zap"
	"net/http"
)

// Router define routes priority
func Router(h *handlers.Handler, db *sql.DB, l *zap.Logger, p *worker.Pool) *mux.Router {
	rtr := mux.NewRouter()
	// Mass save short links
	rtr.HandleFunc("/api/shorten/batch", h.BunchSaveJSON).Methods(http.MethodPost)
	// Save link from JSON format
	rtr.HandleFunc("/api/shorten", h.SaveJSON).Methods(http.MethodPost)
	// Get user session links in JSON
	rtr.HandleFunc("/user/urls", h.GetUrls).Methods(http.MethodGet)
	// Ping db connection
	rtr.Handle("/ping", ping.New(db, l))
	// Delete links session
	rtr.Handle("/api/user/urls", delete.New(db, l, p)).Methods(http.MethodDelete)
	// Get origin by short link
	rtr.HandleFunc("/{id:.+}", h.Get).Methods(http.MethodGet)
	// Save origin to short
	rtr.HandleFunc("/", h.Save).Methods(http.MethodPost)
	return rtr
}

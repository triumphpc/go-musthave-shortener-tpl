package routes

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/delete"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/ping"
	"go.uber.org/zap"
	"net/http"
)

// Router define routes priority
func Router(h *handlers.Handler, db *sql.DB, l *zap.Logger) *mux.Router {
	rtr := mux.NewRouter()
	cookieMw := middlewares.New(l)
	// Mass save short links session
	rtr.Handle(
		"/api/shorten/batch",
		cookieMw.CookieMiddleware(http.HandlerFunc(h.BunchSaveJSON)),
	).Methods(http.MethodPost)
	// Save link from JSON format session
	rtr.Handle(
		"/api/shorten",
		cookieMw.CookieMiddleware(http.HandlerFunc(h.SaveJSON)),
	).Methods(http.MethodPost)
	// Delete links session
	rtr.Handle(
		"/api/user/urls",
		cookieMw.CookieMiddleware(delete.New(db, l)),
	).Methods(http.MethodDelete)
	// Get user session links in JSON
	rtr.Handle(
		"/user/urls",
		cookieMw.CookieMiddleware(http.HandlerFunc(h.GetUrls)),
	).Methods(http.MethodGet)
	// Ping db connection
	rtr.Handle("/ping", ping.New(db, l)).Methods(http.MethodPost)
	// Get origin by short link
	rtr.HandleFunc("/{id:.+}", h.Get).Methods(http.MethodGet)
	// Save origin to short
	rtr.HandleFunc("/", h.Save).Methods(http.MethodPost)
	return rtr
}

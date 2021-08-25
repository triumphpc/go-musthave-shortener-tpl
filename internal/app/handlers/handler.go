package handlers

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storage"
	"io/ioutil"
	"net/http"
)

const Host = "http://localhost:8080"

// ErrBadResponse Package level error
var ErrBadResponse = errors.New("bad request")

// Handler general type for handler
type Handler struct {
	s storage.Repository
}

// New Allocation new handler
func New() *Handler {
	return &Handler{
		s: storage.New(),
	}
}

// Save convert link to shorting and store in database
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Validation
		if r.Body != http.NoBody {
			body, err := ioutil.ReadAll(r.Body)
			if err == nil {
				sl := h.s.Save(string(body))
				// Prepare response
				w.Header().Add("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusCreated)

				slURL := fmt.Sprintf("%s/%s", Host, string(sl))
				_, err = w.Write([]byte(slURL))
				if err == nil {
					return
				}
			}
		}
	}
	setBadResponse(w)
}

// Get fid origin link from storage
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Validation id params
		params := mux.Vars(r)
		id := params["id"]

		if id != "" {
			url, err := h.s.LinkBy(storage.ShortLink(id))
			if err == nil {
				http.Redirect(w, r, url, http.StatusTemporaryRedirect)
				return
			}
		}
	}
	setBadResponse(w)
}

// setBadRequest set bad response
func setBadResponse(w http.ResponseWriter) {
	http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
}

package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storage"
	"io/ioutil"
	"net/http"
)

const Host = "http://localhost:8080"

// Save convert link to shorting and store in database
func Save(s storage.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Validation
			if r.Body != http.NoBody {
				body, err := ioutil.ReadAll(r.Body)
				if err == nil {
					// Save in database
					sl := s.Save(string(body))
					// Prepare response
					w.Header().Add("Content-Type", "text/plain; charset=utf-8")
					w.WriteHeader(http.StatusCreated)

					slUrl := fmt.Sprintf("%s/%s", Host, string(sl))
					_, err = w.Write([]byte(slUrl))
					if err == nil {
						return
					}
				}
			}
		}
		setBadResponse(w)
		return
	}
}

// Get fid origin link from storage
func Get(s storage.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Validation id params
			params := mux.Vars(r)
			id := params["id"]

			if id != "" {
				url, err := s.LinkBy(storage.ShortLink(id))
				if err == nil {
					http.Redirect(w, r, url, http.StatusTemporaryRedirect)
					return
				}
			}
		}
		setBadResponse(w)
		return
	}
}

// setBadRequest set bad response
func setBadResponse(w http.ResponseWriter) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

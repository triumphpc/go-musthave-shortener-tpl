package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/memory"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

// ErrBadResponse Package level error
var ErrBadResponse = errors.New("bad request")
var ErrUnknownURL = errors.New("unknown url")
var ErrInternalError = errors.New("internal error")

// Repository interface for working with global repository
// go:generate mockery --name=Repository --inpackage
type Repository interface {
	// LinkBy get original link
	LinkBy(sl memory.ShortLink) (string, error)
	// Save link to repository
	Save(url string) (sl memory.ShortLink)
}

// Handler general type for handler
type Handler struct {
	s Repository
}

// URL it's users full url
type URL struct {
	URL string `json:"url"`
}

func (h *Handler) SetRepository(r Repository) {
	h.s = r
}

// New Allocation new handler
func New() (h *Handler, err error) {
	// If set file storage path
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return &Handler{
			s: memory.New(),
		}, nil
	} else {
		s, err := file.New()
		if err != nil {
			return nil, err
		}
		return &Handler{
			s: s,
		}, nil
	}
}

// Save convert link to shorting and store in database
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Validation
		if r.Body != http.NoBody {
			body, err := ioutil.ReadAll(r.Body)
			if err == nil {
				origin := string(body)
				short := string(h.s.Save(origin))
				// Prepare response
				w.Header().Add("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusCreated)

				baseURL, err := configs.Instance().Param(configs.BaseURL)
				if err == nil {
					slURL := fmt.Sprintf("%s/%s", baseURL, short)
					_, err = w.Write([]byte(slURL))
					if err == nil {
						return
					}
				}
			}
		}
	}
	setBadResponse(w, ErrBadResponse)
}

// SaveJSON convert link to shorting and store in database
func (h *Handler) SaveJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		setBadResponse(w, ErrBadResponse)
		return

	}
	// Validation
	if r.Body == http.NoBody {
		setBadResponse(w, ErrBadResponse)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		setBadResponse(w, ErrUnknownURL)
		return
	}
	// Get url from json data
	url := URL{}
	err = json.Unmarshal(body, &url)
	if err != nil {
		setBadResponse(w, ErrUnknownURL)
		return
	}

	if url.URL == "" {
		setBadResponse(w, ErrUnknownURL)
		return
	}

	sl := h.s.Save(url.URL)
	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		setBadResponse(w, ErrBadResponse)
	}
	slURL := fmt.Sprintf("%s/%s", baseURL, string(sl))
	result := struct {
		Result string `json:"result"`
	}{Result: slURL}

	// log to stdout
	logger.Info("save to json format", zap.Reflect("URL", result))

	body, err = json.Marshal(result)
	if err == nil {
		// Prepare response
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(body)
		if err == nil {
			return
		}
	}
	setBadResponse(w, ErrInternalError)
}

// Get fid origin link from storage
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Validation id params
		params := mux.Vars(r)
		id := params["id"]

		if id != "" {
			url, err := h.s.LinkBy(memory.ShortLink(id))
			if err == nil {
				http.Redirect(w, r, url, http.StatusTemporaryRedirect)
				return
			} else {
				logger.Info("Get error", zap.Error(err))
			}
		}
	}
	setBadResponse(w, ErrBadResponse)
}

// setBadRequest set bad response
func setBadResponse(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusBadRequest)
}

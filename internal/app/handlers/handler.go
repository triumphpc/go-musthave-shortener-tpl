package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
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
	// LinkByShort get original link
	LinkByShort(userId user.UniqUser, short shortlink.Short) (string, error)
	// Save link to repository
	Save(userId user.UniqUser, url string) shortlink.Short
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
	s, err := file.New()
	if err != nil {
		return nil, err
	}
	return &Handler{
		s: s,
	}, nil
}

// Save convert link to shorting and store in database
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Validation
		if r.Body != http.NoBody {
			body, err := ioutil.ReadAll(r.Body)
			if err == nil {
				origin := string(body)
				// Get userId from context
				userIdCtx := r.Context().Value(middlewares.UserIdCtxName)
				// Convert interface type to user.UniqUser
				userId := userIdCtx.(string)
				short := string(h.s.Save(user.UniqUser(userId), origin))
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
	userIdCtx := r.Context().Value(middlewares.UserIdCtxName)
	// Convert interface type to user.UniqUser
	userId := userIdCtx.(string)
	sl := h.s.Save(user.UniqUser(userId), url.URL)

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
			userIdCtx := r.Context().Value(middlewares.UserIdCtxName)
			// Convert interface type to user.UniqUser
			userId := userIdCtx.(string)
			url, err := h.s.LinkByShort(user.UniqUser(userId), shortlink.Short(id))
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

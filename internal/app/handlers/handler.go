package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	dbh "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

// ErrBadResponse Package level error
var ErrBadResponse = errors.New("bad request")
var ErrUnknownURL = errors.New("unknown url")
var ErrInternalError = errors.New("internal error")
var ErrNoContent = errors.New("no content")

// Repository interface for working with global repository
// go:generate mockery --name=Repository --inpackage
type Repository interface {
	// LinkByShort get original link from all storage
	LinkByShort(short shortlink.Short) (string, error)
	// Save link to repository
	Save(userID user.UniqUser, url string) shortlink.Short
	// BunchSave save mass urls and generate shorts
	BunchSave(urls []shortlink.URLs) ([]shortlink.ShortURLs, error)
	// LinksByUser return all user links
	LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error)
}

// Handler general type for handler
type Handler struct {
	s Repository
}

// New Allocation new handler
func New() (*Handler, error) {
	// Check in db has
	_, err := db.Instance()
	if err == nil {
		logger.Info("Set db handler")
		s, err := dbh.New()
		if err != nil {
			return nil, err
		}
		return &Handler{
			s: s,
		}, nil
	} else {
		logger.Info("Set file handler")
		// File and memory storage
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
	if r.Body != http.NoBody {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			origin := string(body)
			// Get userID from context
			userIDCtx := r.Context().Value(middlewares.UserIDCtxName)
			userID := "default"
			if userIDCtx != nil {
				// Convert interface type to user.UniqUser
				userID = userIDCtx.(string)
			}
			short := string(h.s.Save(user.UniqUser(userID), origin))
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
	setBadResponse(w, ErrBadResponse)
}

// SaveJSON convert link to shorting and store in database
func (h *Handler) SaveJSON(w http.ResponseWriter, r *http.Request) {
	body, err := bodyFromJSON(&w, r)
	if err != nil {
		setBadResponse(w, ErrInternalError)
		return
	}
	// Get url from json data
	url := shortlink.URL{}
	err = json.Unmarshal(body, &url)

	if err != nil {
		setBadResponse(w, ErrUnknownURL)
		return
	}
	if url.URL == "" {
		setBadResponse(w, ErrUnknownURL)
		return
	}
	userIDCtx := r.Context().Value(middlewares.UserIDCtxName)
	userID := "default"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}
	sl := h.s.Save(user.UniqUser(userID), url.URL)

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

// BunchSaveJSON save data and return from mass
func (h *Handler) BunchSaveJSON(w http.ResponseWriter, r *http.Request) {
	body, err := bodyFromJSON(&w, r)
	if err != nil {
		setBadResponse(w, ErrInternalError)
		return
	}
	// Get url from json data
	var urls []shortlink.URLs
	err = json.Unmarshal(body, &urls)
	if err != nil {
		setBadResponse(w, ErrUnknownURL)
		return
	}
	shorts, err := h.s.BunchSave(urls)
	if err != nil {
		setBadResponse(w, ErrInternalError)
		return
	}
	// Determine base url
	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		setBadResponse(w, ErrBadResponse)
	}
	// Prepare results
	for k := range shorts {
		shorts[k].Short = fmt.Sprintf("%s/%s", baseURL, shorts[k].Short)
	}

	body, err = json.Marshal(shorts)
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

// bodyFromJSON get bytes from JSON requests
func bodyFromJSON(w *http.ResponseWriter, r *http.Request) ([]byte, error) {
	var body []byte
	if r.Body == http.NoBody {
		setBadResponse(*w, ErrBadResponse)
		return body, ErrBadResponse
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		setBadResponse(*w, ErrUnknownURL)
		return body, ErrUnknownURL
	}
	return body, nil
}

// Get fid origin link from storage
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Validation id params
	params := mux.Vars(r)
	id := params["id"]
	fmt.Print(id)
	if id != "" {
		url, err := h.s.LinkByShort(shortlink.Short(id))
		if err == nil {
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		} else {
			logger.Info("Get error", zap.Error(err))
		}
	}
	setBadResponse(w, ErrBadResponse)
}

// GetUrls all urls from user
func (h *Handler) GetUrls(w http.ResponseWriter, r *http.Request) {
	userIDCtx := r.Context().Value(middlewares.UserIDCtxName)
	// Convert interface type to user.UniqUser
	userID := userIDCtx.(string)
	links, err := h.s.LinksByUser(user.UniqUser(userID))
	if err != nil {
		http.Error(w, ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	type coupleLinks struct {
		Short  string `json:"short_url"`
		Origin string `json:"original_url"`
	}
	var lks []coupleLinks
	baseURL, _ := configs.Instance().Param(configs.BaseURL)

	// Get all links
	for k, v := range links {
		lks = append(lks, coupleLinks{
			Short:  fmt.Sprintf("%s/%s", baseURL, string(k)),
			Origin: v,
		})
	}
	body, err := json.Marshal(lks)
	if err == nil {
		// Prepare response
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(body)
		if err == nil {
			return
		}
	}
	setBadResponse(w, ErrBadResponse)
}

// setBadRequest set bad response
func setBadResponse(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusBadRequest)
}

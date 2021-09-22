package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
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
	Save(userID user.UniqUser, url string) (shortlink.Short, error)
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
func New(c *sql.DB) (*Handler, error) {
	// Check in db has
	if c != nil {
		logger.Info("Set db handler")
		s, err := dbh.New(c)
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
	if r.Body == http.NoBody {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return

	}
	origin := string(body)
	// Get userID from context
	userIDCtx := r.Context().Value(middlewares.UserIDCtxName)
	userID := "default"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}
	short, err := h.s.Save(user.UniqUser(userID), origin)
	status := http.StatusCreated
	if errors.Is(err, dbh.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}
	// Prepare response
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	slURL := fmt.Sprintf("%s/%s", baseURL, short)
	_, err = w.Write([]byte(slURL))
	if err != nil {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
	}
}

// SaveJSON convert link to shorting and store in database
func (h *Handler) SaveJSON(w http.ResponseWriter, r *http.Request) {
	body, err := bodyFromJSON(&w, r)
	if err != nil {
		http.Error(w, ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Get url from json data
	url := shortlink.URL{}
	err = json.Unmarshal(body, &url)

	if err != nil {
		http.Error(w, ErrUnknownURL.Error(), http.StatusBadRequest)
		return
	}
	if url.URL == "" {
		http.Error(w, ErrUnknownURL.Error(), http.StatusBadRequest)
		return
	}
	userIDCtx := r.Context().Value(middlewares.UserIDCtxName)
	userID := "default"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}
	short, err := h.s.Save(user.UniqUser(userID), url.URL)
	status := http.StatusCreated
	if errors.Is(err, dbh.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}
	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	slURL := fmt.Sprintf("%s/%s", baseURL, string(short))
	result := struct {
		Result string `json:"result"`
	}{Result: slURL}

	// log to stdout
	logger.Info("save to json format", zap.Reflect("URL", result))
	body, err = json.Marshal(result)
	if err != nil {
		http.Error(w, ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, ErrInternalError.Error(), http.StatusBadRequest)
	}
}

// BunchSaveJSON save data and return from mass
func (h *Handler) BunchSaveJSON(w http.ResponseWriter, r *http.Request) {
	body, err := bodyFromJSON(&w, r)
	if err != nil {
		http.Error(w, ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Get url from json data
	var urls []shortlink.URLs
	err = json.Unmarshal(body, &urls)
	if err != nil {
		http.Error(w, ErrUnknownURL.Error(), http.StatusBadRequest)
		return
	}
	shorts, err := h.s.BunchSave(urls)
	if err != nil {
		http.Error(w, ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Determine base url
	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	// Prepare results
	for k := range shorts {
		shorts[k].Short = fmt.Sprintf("%s/%s", baseURL, shorts[k].Short)
	}
	body, err = json.Marshal(shorts)
	if err != nil {
		http.Error(w, ErrInternalError.Error(), http.StatusBadRequest)
		return

	}
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, ErrInternalError.Error(), http.StatusBadRequest)
	}
}

// bodyFromJSON get bytes from JSON requests
func bodyFromJSON(w *http.ResponseWriter, r *http.Request) ([]byte, error) {
	var body []byte
	if r.Body == http.NoBody {
		http.Error(*w, ErrBadResponse.Error(), http.StatusBadRequest)
		return body, ErrBadResponse
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(*w, ErrUnknownURL.Error(), http.StatusBadRequest)
		return body, ErrUnknownURL
	}
	return body, nil
}

// Get fid origin link from storage
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Validation id params
	params := mux.Vars(r)
	id := params["id"]

	if id == "" {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	url, err := h.s.LinkByShort(shortlink.Short(id))
	if err != nil {
		logger.Info("Get error", zap.Error(err))
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
	if err != nil {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, ErrBadResponse.Error(), http.StatusBadRequest)
	}
}

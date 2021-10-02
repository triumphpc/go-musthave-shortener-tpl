package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	dbh "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

// Repository interface for working with global repository
// go:generate mockery --name=Repository --inpackage
type Repository interface {
	// LinkByShort get original link from all storage
	LinkByShort(short shortlink.Short, userID user.UniqUser) (string, error)
	// Save link to repository
	Save(userID user.UniqUser, url string) (shortlink.Short, error)
	// BunchSave save mass urls and generate shorts
	BunchSave(urls []shortlink.URLs, userID user.UniqUser) ([]shortlink.ShortURLs, error)
	// LinksByUser return all user links
	LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error)
}

// Handler general type for handler
type Handler struct {
	s Repository
	l *zap.Logger
}

// New Allocation new handler
func New(c *sql.DB, l *zap.Logger) (*Handler, error) {
	var s Repository
	var err error

	// Check in db has
	if c != nil {
		l.Info("Set db handler")
		s, err = dbh.New(c, l)
	} else {
		l.Info("Set file handler")
		// File and memory storage
		s, err = file.New()
	}

	if err != nil {
		return nil, err
	}

	return &Handler{
		s: s,
		l: l,
	}, nil
}

// Save convert link to shorting and store in database
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return

	}
	origin := string(body)
	short, err := h.s.Save(helpers.GetContextUserID(r), origin)
	status := http.StatusCreated
	if errors.Is(err, er.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}
	// Prepare response
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	slURL := fmt.Sprintf("%s/%s", baseURL, short)
	_, err = w.Write([]byte(slURL))
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
	}
}

// SaveJSON convert link to shorting and store in database
func (h *Handler) SaveJSON(w http.ResponseWriter, r *http.Request) {
	body, err := helpers.BodyFromJSON(&w, r)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Get url from json data
	url := shortlink.URL{}
	err = json.Unmarshal(body, &url)

	if err != nil {
		http.Error(w, er.ErrUnknownURL.Error(), http.StatusBadRequest)
		return
	}
	if url.URL == "" {
		http.Error(w, er.ErrUnknownURL.Error(), http.StatusBadRequest)
		return
	}

	short, err := h.s.Save(helpers.GetContextUserID(r), url.URL)
	status := http.StatusCreated
	if errors.Is(err, er.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}
	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	slURL := fmt.Sprintf("%s/%s", baseURL, string(short))
	result := struct {
		Result string `json:"result"`
	}{Result: slURL}

	// log to stdout
	h.l.Info("save to json format", zap.Reflect("URL", result))
	body, err = json.Marshal(result)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
	}
}

// BunchSaveJSON save data and return from mass
func (h *Handler) BunchSaveJSON(w http.ResponseWriter, r *http.Request) {
	body, err := helpers.BodyFromJSON(&w, r)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Get url from json data
	var urls []shortlink.URLs
	err = json.Unmarshal(body, &urls)
	if err != nil {
		http.Error(w, er.ErrUnknownURL.Error(), http.StatusBadRequest)
		return
	}
	shorts, err := h.s.BunchSave(urls, helpers.GetContextUserID(r))
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}
	// Determine base url
	baseURL, err := configs.Instance().Param(configs.BaseURL)
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	// Prepare results
	for k := range shorts {
		shorts[k].Short = fmt.Sprintf("%s/%s", baseURL, shorts[k].Short)
	}
	body, err = json.Marshal(shorts)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return

	}
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
	}
}

// Get fid origin link from storage
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Validation id params
	params := mux.Vars(r)
	id := params["id"]

	if id == "" {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	url, err := h.s.LinkByShort(shortlink.Short(id), helpers.GetContextUserID(r))
	if err != nil {
		h.l.Info("Get error", zap.Error(err))
		if errors.Is(err, er.ErrURLIsGone) {
			http.Error(w, er.ErrURLIsGone.Error(), http.StatusGone)
			return
		}

		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GetUrls all urls from user
func (h *Handler) GetUrls(w http.ResponseWriter, r *http.Request) {
	links, err := h.s.LinksByUser(helpers.GetContextUserID(r))
	if err != nil {
		http.Error(w, er.ErrNoContent.Error(), http.StatusNoContent)
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
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
	}
}

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/repository"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

// Repository interface for working with global repository
// go:generate mockery --name=Repository --inpackage
//type Repository interface {
//	// LinkByShort get original link from all storage
//	LinkByShort(short shortlink.Short) (string, error)
//	// Save link to repository
//	Save(userID user.UniqUser, url string) (shortlink.Short, error)
//	// BunchSave save mass urls and generate shorts
//	BunchSave(userID user.UniqUser, urls []shortlink.URLs) ([]shortlink.ShortURLs, error)
//	// LinksByUser return all user links
//	LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error)
//	// Clear storage
//	Clear() error
//	// BunchUpdateAsDeleted set flag as deleted
//	BunchUpdateAsDeleted(ctx context.Context, ids []string, userID string) error
//}

// Handler general type for handler
type Handler struct {
	s repository.Repository
	l *zap.Logger
}

// New Allocation new handler
func New(l *zap.Logger, s repository.Repository) *Handler {
	return &Handler{s, l}
}

// Save convert link to shorting and store in database
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	h.l.Info("Save run")
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

	h.l.Info("Save origin", zap.String("origin", origin))

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
	h.l.Info("SaveJSON run")
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

	h.l.Info("save origin", zap.String("URL", url.URL))

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

	h.l.Info("save to json format", zap.Reflect("URL", result))
	body, err = json.Marshal(result)
	if err != nil {
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
		return
	}

	h.l.Info("prepare response")
	// Prepare response
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(body)
	if err != nil {
		h.l.Info("base status request")
		http.Error(w, er.ErrInternalError.Error(), http.StatusBadRequest)
	}
}

// BunchSaveJSON save data and return from mass
func (h *Handler) BunchSaveJSON(w http.ResponseWriter, r *http.Request) {
	h.l.Info("BunchSaveJSON run")
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

	shorts, err := h.s.BunchSave(helpers.GetContextUserID(r), urls)
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
	h.l.Info("Get run")
	// Validation id params
	params := mux.Vars(r)
	id := params["id"]

	if id == "" {
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	h.l.Info("Get id:", zap.String("id", id))
	url, err := h.s.LinkByShort(shortlink.Short(id))

	h.l.Info("Result err", zap.Error(err))
	h.l.Info("Result url", zap.String("url", url))

	if err != nil {
		if errors.Is(err, er.ErrURLIsGone) {
			h.l.Info("Get error is gone", zap.Error(err))
			http.Error(w, er.ErrURLIsGone.Error(), http.StatusGone)
			return
		}

		h.l.Info("Get error", zap.Error(err))
		http.Error(w, er.ErrBadResponse.Error(), http.StatusBadRequest)
		return
	}
	h.l.Info("redirect")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GetUrls all urls from user
func (h *Handler) GetUrls(w http.ResponseWriter, r *http.Request) {
	h.l.Info("GetUrls run")
	links, err := h.s.LinksByUser(helpers.GetContextUserID(r))

	if err != nil || len(links) == 0 {
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
	h.l.Info("User links", zap.Reflect("links", lks))

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

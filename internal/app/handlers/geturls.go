package handlers

import (
	"net/http"
)

type GetUrlsHandler struct {
	// Main handler
	H *Handler
}

func (h GetUrlsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.H.GetUrls(w, r)
}

package handlers

import "net/http"

type SaveHandler struct {
	// Main handler
	H *Handler
}

func (h SaveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.H.Save(w, r)
}

package handlers

import (
	"net/http"
)

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method must be GET", http.StatusMethodNotAllowed)
		return
	}
	pong := []byte("Pong")
	w.WriteHeader(http.StatusOK)
	w.Write(pong)
}

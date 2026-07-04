package handlefuncs

import (
	"net/http"
)

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
	serviceName := r.URL.Query().Get("service")
	h.Core.Add(serviceName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

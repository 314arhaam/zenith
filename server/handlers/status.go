package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	serviceName := r.URL.Query().Get("service")
	w.Header().Set("Content-Type", "application/json")
	if serviceName != "" {
		val, ok := h.Core.Get(serviceName)
		if !ok {
			msg := strings.Join([]string{"Service Not Found: ", serviceName}, "")
			http.Error(w, msg, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(val); err != nil {
			http.Error(w, "Error in encoding", http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(h.Core.GetAll()); err != nil {
			http.Error(w, "Error in encoding", http.StatusInternalServerError)
		}
	}
}

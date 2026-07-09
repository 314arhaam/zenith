package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"zenith/core"
)

func (h *Handler) status(serviceName string) (map[string]core.Service, error) {
	if serviceName != "" {
		val, ok := h.Core.Get(serviceName)
		if !ok {
			return nil, fmt.Errorf("Service not found: %s", serviceName)
		}
		return map[string]core.Service{serviceName: val}, nil
	} else {
		val := h.Core.GetAll()
		return val, nil
	}
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	serviceName := r.URL.Query().Get("service")
	w.Header().Set("Content-Type", "application/json")
	val, err := h.status(serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(val); err != nil {
		http.Error(w, "Error in encoding", http.StatusInternalServerError)
		return
	}
}

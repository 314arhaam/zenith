package handlers

import (
	"net/http"
	data "zenith/models"
)

func (h *Handler) remove(serviceName string) bool {
	return h.Core.Remove(serviceName)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	var rs data.RemoveRequest
	if err := data.Decode(&rs, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ok := rs.Validate(); !ok {
		http.Error(w, "No Data Provided", http.StatusBadRequest)
		return
	}
	if ok := h.remove(rs.ServiceName); !ok {
		http.Error(w, "Element not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

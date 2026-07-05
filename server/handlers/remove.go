package handlers

import (
	"encoding/json"
	"net/http"
	data "zenith/models"
)

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	var rs data.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&rs); err != nil {
		http.Error(w, "Error in request decode "+err.Error(), http.StatusInternalServerError)
		return
	}
	if rs.ServiceName == "" {
		http.Error(w, "Not Data Provided", http.StatusBadRequest)
		return
	}
	h.Core.Remove(rs.ServiceName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

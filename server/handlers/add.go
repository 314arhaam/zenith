package handlers

import (
	"encoding/json"
	"net/http"
	"zenith/core"
	data "zenith/models"
)

func (h *Handler) add(serviceName string) (core.Service, bool) {
	h.Core.Add(serviceName)
	val, ok := h.Core.Get(serviceName)
	return val, ok
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	// check if method ok
	// Post -> ok
	// other -> http.StatusMethodNotAllowed -> return
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method", http.StatusMethodNotAllowed)
		return
	}
	// read request data
	defer r.Body.Close()
	var rs data.AddRequest
	if err := data.Decode(&rs, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// validate request, check if it has the key
	if ok := rs.Validate(); !ok {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	// core logic
	val, ok := h.add(rs.ServiceName)
	if !ok {
		http.Error(w, "Error in data check", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(val); err != nil {
		http.Error(w, "Error in encoding", http.StatusInternalServerError)
		return
	}
}

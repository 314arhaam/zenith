package handlers

import (
	"encoding/json"
	"net/http"
	data "zenith/models"
)

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	// check if method ok
	// Post -> ok
	// other -> http.StatusMethodNotAllowed -> return
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	// read request data
	defer r.Body.Close()
	var rs data.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&rs); err != nil {
		http.Error(w, "Error in request decode "+err.Error(), http.StatusInternalServerError)
		return
	}
	// validate request, check if it has the key
	if rs.ServiceName == "" {
		http.Error(w, "Not Data Provided", http.StatusBadRequest)
		return
	}
	// core logic
	h.Core.Add(rs.ServiceName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	val, ok := h.Core.Get(rs.ServiceName)
	if !ok {
		http.Error(w, "Error in data check", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(val); err != nil {
		http.Error(w, "Error in encoding", http.StatusInternalServerError)
		return
	}
}

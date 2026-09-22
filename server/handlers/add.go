package handlers

import (
	"net/http"
	data "zenith/models"
)

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	var request data.AddRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	service, created := h.Core.Add(request.ServiceName)
	if !created {
		writeError(w, http.StatusConflict, "service_exists", "service already exists")
		return
	}
	writeJSON(w, http.StatusCreated, service)
}

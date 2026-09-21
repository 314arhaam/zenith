package handlers

import (
	"net/http"
	data "zenith/models"
)

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w, "DELETE")
		return
	}
	var request data.RemoveRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if !h.Core.Remove(request.ServiceName) {
		writeError(w, http.StatusNotFound, "service_not_found", "service not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

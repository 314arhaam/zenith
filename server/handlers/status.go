package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"zenith/core"
	data "zenith/models"
)

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, "GET, HEAD")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", "invalid query encoding")
		return
	}
	if names, provided := query["service"]; provided {
		if len(names) != 1 {
			writeError(w, http.StatusBadRequest, "invalid_query", "provide service once")
			return
		}
		name, valid := data.NormalizeServiceName(names[0])
		if !valid {
			writeError(w, http.StatusBadRequest, "invalid_query", "invalid service name")
			return
		}
		service, ok := h.Core.Get(name)
		if !ok {
			if r.Method == http.MethodHead {
				w.Header().Set("X-Zenith-Service-Count", "0")
				w.WriteHeader(http.StatusNotFound)
			} else {
				writeError(w, http.StatusNotFound, "service_not_found", fmt.Sprintf("service not found: %s", name))
			}
			return
		}
		if r.Method == http.MethodHead {
			w.Header().Set("X-Zenith-Service-Count", "1")
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, map[string]core.Service{name: service})
		return
	}
	if r.Method == http.MethodHead {
		w.Header().Set("X-Zenith-Service-Count", strconv.Itoa(h.Core.Len()))
		w.WriteHeader(http.StatusOK)
		return
	}
	body, err := h.Core.Marshal()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encode_error", "cannot encode status")
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, body)
	_, _ = io.WriteString(w, "\n")
}

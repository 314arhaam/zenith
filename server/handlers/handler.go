package handlers

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"zenith/core"
	data "zenith/models"
)

// MaxRequestBytes bounds each registration/removal payload, not total registry
// size. This is intentionally much larger than the maximum service name.
const MaxRequestBytes int64 = 4096

type Handler struct{ Core *core.System }

func NewHandler() *Handler { return &Handler{Core: core.NewSystem()} }

// Routes is shared by the real server and integration tests to prevent drift.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/add", h.Add)
	mux.HandleFunc("/remove", h.Remove)
	mux.HandleFunc("/status", h.Status)
	mux.HandleFunc("/ping", h.Ping)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found")
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, data.ErrorResponse{Error: code, Message: message})
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use "+allowed+" method")
}

func decodeRequest(w http.ResponseWriter, r *http.Request, payload data.Request) bool {
	defer r.Body.Close()
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "use application/json")
			return false
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBytes)
	if err := data.Decode(payload, r.Body); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "request body exceeds 4096 bytes")
		} else {
			writeError(w, http.StatusBadRequest, "invalid_json", "expected one JSON object with service_name")
		}
		return false
	}
	if !payload.Validate() {
		writeError(w, http.StatusBadRequest, "invalid_payload", "service_name must be 1-256 UTF-8 bytes without control characters")
		return false
	}
	return true
}

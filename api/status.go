package handlefuncs

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
	serviceName := r.URL.Query().Get("service")
	var jsonData []byte
	var err error
	if serviceName != "" {
		val, ok := h.Core.Get(serviceName)
		if !ok {
			http.Error(w, "Service Not Found", http.StatusNotFound)
			return
		}
		jsonData, err = json.MarshalIndent(val, "", " ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		jsonData, err = json.MarshalIndent(h.Core.GetAll(), "", " ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
	w.WriteHeader(http.StatusOK)
}

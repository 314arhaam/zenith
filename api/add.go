package handlefuncs

import (
	"net/http"
	data "zenith/models"
)

func Add(w http.ResponseWriter, r *http.Request, data *data.ServiceData) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
	serviceName := r.URL.Query().Get("service")
	data.Add(serviceName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

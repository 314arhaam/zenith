package handlefuncs

import (
	"net/http"
	data "zenith/models"
)

func Remove(w http.ResponseWriter, r *http.Request, data *data.ServiceData) {
	serviceName := r.URL.Query().Get("service")
	data.Remove(serviceName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

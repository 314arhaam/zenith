package handlefuncs

import (
	"net/http"
	data "zenith/models"
)

func Add(w http.ResponseWriter, r *http.Request, data *data.ServiceData) {
	serviceName := r.URL.Query().Get("service")
	data.Add(serviceName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

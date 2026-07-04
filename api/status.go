package handlefuncs

import (
	"encoding/json"
	"net/http"
	data "zenith/models"
)

func Status(w http.ResponseWriter, r *http.Request, data *data.ServiceData) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
	jsonData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
	w.WriteHeader(http.StatusCreated)
}

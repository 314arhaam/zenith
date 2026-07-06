package data

import "encoding/json"

type RequestPayload struct {
	ServiceName string `json:"service_name"`
}

func (r *RequestPayload) ToJson() ([]byte, error) {
	return json.Marshal(r)
}

package data

import (
	"encoding/json"
	"io"
)

type AddRequest struct {
	ServiceName string `json:"service_name"`
}

func (r AddRequest) Validate() bool {
	if r.ServiceName == "" {
		return false
	}
	return true
}

type RemoveRequest struct {
	ServiceName string `json:"service_name"`
}

func (r RemoveRequest) Validate() bool {
	if r.ServiceName == "" {
		return false
	}
	return true
}

type Request interface {
	Validate() bool
}

func Decode(r Request, body io.ReadCloser) error {
	if err := json.NewDecoder(body).Decode(r); err != nil {
		return err
	}
	return nil
}

func ToJson(r Request) (string, error) {
	val, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

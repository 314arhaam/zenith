package data

import (
	"encoding/json"
	"io"
)

type Request interface {
	Validate() bool
}

func Decode(r Request, body io.ReadCloser) error {
	if err := json.NewDecoder(body).Decode(&r); err != nil {
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

package data

import (
	"encoding/json"
	"fmt"
	"io"
)

type Request interface {
	Validate() bool
}

func Decode(r Request, body io.Reader) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(r); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("request body must contain a single JSON object")
		}
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

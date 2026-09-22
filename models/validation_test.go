package data

import (
	"strings"
	"testing"
)

func TestNameValidation(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"api", true}, {"  api  ", true}, {"api & worker/+?=", true}, {"\u062e\u062f\u0645\u062a", true},
		{"", false}, {" \t\n", false}, {"api\nworker", false}, {"a\x00b", false}, {"a\x7fb", false},
		{strings.Repeat("x", 256), true}, {strings.Repeat("x", 257), false}, {"\xff", false},
	}
	for _, tt := range tests {
		name, ok := NormalizeServiceName(tt.name)
		if ok != tt.valid {
			t.Errorf("NormalizeServiceName(%q) = %q, %v", tt.name, name, ok)
		}
		for _, req := range []Request{&AddRequest{ServiceName: tt.name}, &RemoveRequest{ServiceName: tt.name}} {
			if req.Validate() != tt.valid {
				t.Errorf("Validate(%q)", tt.name)
			}
		}
	}
}

func TestStrictJSON(t *testing.T) {
	for _, body := range []string{"", "{", "[]", `{"service_name":5}`, `{"service_name":"api","typo":1}`, `{"service_name":"api"}null`, `{"service_name":"api"} garbage`} {
		var request AddRequest
		if err := Decode(&request, strings.NewReader(body)); err == nil {
			t.Errorf("accepted invalid JSON %q", body)
		}
	}
	for _, body := range []string{`{}`, `null`, `{"service_name":null}`} {
		var request AddRequest
		if err := Decode(&request, strings.NewReader(body)); err == nil && request.Validate() {
			t.Errorf("accepted missing name %q", body)
		}
	}
}

func FuzzDecodeRequest(f *testing.F) {
	for _, seed := range []string{`{"service_name":"api"}`, `null`, `{}`, `[]`, `{"service_name":"a"} {}`, "\xff"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > 8192 {
			return
		}
		var request AddRequest
		if Decode(&request, strings.NewReader(body)) == nil && request.Validate() {
			encoded, err := ToJson(&request)
			if err != nil {
				t.Fatal(err)
			}
			var roundtrip AddRequest
			if err := Decode(&roundtrip, strings.NewReader(encoded)); err != nil || roundtrip.ServiceName != request.ServiceName {
				t.Fatal("roundtrip failed")
			}
		}
	})
}

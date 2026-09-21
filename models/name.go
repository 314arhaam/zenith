package data

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const MaxServiceNameBytes = 256

// NormalizeServiceName is shared by all HTTP and CLI entry points. Interior
// spaces, Unicode and URL-special characters are valid; controls are not.
func NormalizeServiceName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > MaxServiceNameBytes || !utf8.ValidString(name) {
		return "", false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", false
		}
	}
	return name, true
}

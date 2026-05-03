package selector

import (
	"strings"
)

func IsCSS(s string) bool {
	if IsRef(s) {
		return false
	}
	prefixes := []string{"#", ".", "[", "*", ":", ">", "+", "~", " "}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	for _, c := range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" {
		if s == string(c) || strings.HasPrefix(s, string(c)+":") || strings.HasPrefix(s, string(c)+">") {
			return true
		}
	}
	return true
}

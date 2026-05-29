package util

import "strings"

func DeriveClientName(email string) string {
	at := strings.IndexByte(email, '@')
	if at == -1 {
		return "usuario"
	}
	prefix := email[:at]
	var clean strings.Builder
	for _, r := range prefix {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			clean.WriteRune(r)
		}
	}
	name := clean.String()
	if name == "" {
		return "usuario"
	}
	return "usuario-" + name
}

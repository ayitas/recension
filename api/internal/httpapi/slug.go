package httpapi

import (
	"regexp"
	"strings"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func normalizeSlug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func validSlug(s string) bool {
	return len(s) >= 2 && len(s) <= 64 && slugRe.MatchString(s)
}

package search_engine

import (
	"regexp"
	"strings"
)

var tokenRE = regexp.MustCompile(`[a-z0-9][a-z0-9+#.-]*`)

func Tokens(text string) []string {
	matches := tokenRE.FindAllString(strings.ToLower(text), -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		match = strings.Trim(match, ".-")
		if match != "" {
			out = append(out, match)
		}
	}
	return out
}

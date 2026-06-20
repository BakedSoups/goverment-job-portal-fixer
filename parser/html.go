package parser

import (
	"html"
	"regexp"
	"strings"
)

var (
	tagBreakRE = regexp.MustCompile(`(?i)</?(p|div|br|li|ul|ol|h[1-6])[^>]*>`)
	tagRE      = regexp.MustCompile(`<[^>]+>`)
	spaceRE    = regexp.MustCompile(`[ \t\r\n]+`)
)

func HTMLToText(input string) string {
	text := tagBreakRE.ReplaceAllString(input, "\n")
	text = tagRE.ReplaceAllString(text, " ")
	text = html.UnescapeString(text)
	return CleanText(text)
}

func CleanText(input string) string {
	input = strings.ReplaceAll(input, "\u00a0", " ")
	lines := strings.Split(input, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(spaceRE.ReplaceAllString(line, " "))
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

package search_engine

import (
	"regexp"
	"strings"

	"github.com/BakedSoups/goverment-job-portal-fixer/taxonomy"
)

type Matcher struct {
	concepts []compiledConcept
}

type compiledConcept struct {
	concept taxonomy.Concept
	regexps []*regexp.Regexp
}

func NewMatcher() Matcher {
	var concepts []compiledConcept
	for _, concept := range taxonomy.Concepts() {
		compiled := compiledConcept{concept: concept}
		aliases := append([]string{concept.Name, concept.Label}, concept.Aliases...)
		for _, alias := range aliases {
			if isQueryOnlyAlias(alias) {
				continue
			}
			pattern := `(?i)\b` + regexp.QuoteMeta(strings.ToLower(alias)) + `\b`
			compiled.regexps = append(compiled.regexps, regexp.MustCompile(pattern))
		}
		concepts = append(concepts, compiled)
	}
	return Matcher{concepts: concepts}
}

func (m Matcher) Match(text string) map[string]int {
	out := make(map[string]int)
	lower := strings.ToLower(text)
	for _, compiled := range m.concepts {
		for _, re := range compiled.regexps {
			out[compiled.concept.Name] += len(re.FindAllStringIndex(lower, -1))
		}
	}
	return out
}

func isQueryOnlyAlias(alias string) bool {
	switch strings.ToLower(strings.TrimSpace(alias)) {
	case "administration",
		"administrative",
		"analytics",
		"budget",
		"finance",
		"leadership",
		"management",
		"manager",
		"operations",
		"programs":
		return true
	default:
		return false
	}
}

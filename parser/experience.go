package parser

import (
	"regexp"
	"strconv"
	"strings"
)

type ExperienceRequirement struct {
	Min        int
	Max        int
	Found      bool
	Source     string
	Confidence string
}

var (
	yearRangeRE = regexp.MustCompile(`(?i)\b([0-9]+|one|two|three|four|five|six|seven|eight|nine|ten)\s*(?:-|–|to)\s*([0-9]+|one|two|three|four|five|six|seven|eight|nine|ten)\s+years?\b`)
	yearRE      = regexp.MustCompile(`(?i)\b([0-9]+|one|two|three|four|five|six|seven|eight|nine|ten)\s*(?:\([^)]+\)\s*)?years?\b`)
)

func RequiredExperience(text string) ExperienceRequirement {
	lines := strings.Split(CleanText(text), "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if isPreferredExperienceSection(lower) {
			break
		}
		if shouldIgnoreExperienceLine(lower) {
			continue
		}
		if !strings.Contains(lower, "experience") {
			continue
		}

		if min, max, ok := yearsFromLine(line); ok {
			return ExperienceRequirement{
				Min:        min,
				Max:        max,
				Found:      true,
				Source:     line,
				Confidence: "high",
			}
		}
	}

	return ExperienceRequirement{}
}

func PreferredExperience(text string) ExperienceRequirement {
	lines := strings.Split(CleanText(text), "\n")
	inPreferred := false
	for _, line := range lines {
		lower := strings.ToLower(line)
		if isPreferredExperienceSection(lower) {
			inPreferred = true
			continue
		}
		if !inPreferred || shouldIgnoreExperienceLine(lower) || !strings.Contains(lower, "experience") {
			continue
		}
		if min, max, ok := yearsFromLine(line); ok {
			return ExperienceRequirement{
				Min:        min,
				Max:        max,
				Found:      true,
				Source:     line,
				Confidence: "medium",
			}
		}
	}
	return ExperienceRequirement{}
}

func yearsFromLine(line string) (int, int, bool) {
	if match := yearRangeRE.FindStringSubmatch(line); len(match) == 3 {
		min, minOK := numberWord(match[1])
		max, maxOK := numberWord(match[2])
		return min, max, minOK && maxOK
	}
	if match := yearRE.FindStringSubmatch(line); len(match) == 2 {
		min, ok := numberWord(match[1])
		return min, min, ok
	}
	return 0, 0, false
}

func numberWord(raw string) (int, bool) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if n, err := strconv.Atoi(raw); err == nil {
		return n, true
	}
	words := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	}
	n, ok := words[raw]
	return n, ok
}

func isPreferredExperienceSection(line string) bool {
	return strings.Contains(line, "preferred qualifications") ||
		strings.Contains(line, "desirable qualifications")
}

func shouldIgnoreExperienceLine(line string) bool {
	ignored := []string{
		"substitution",
		"may be substituted",
		"equivalent to",
		"full-time employment is equivalent",
		"qualifying work experience is based",
	}
	for _, phrase := range ignored {
		if strings.Contains(line, phrase) {
			return true
		}
	}
	return false
}

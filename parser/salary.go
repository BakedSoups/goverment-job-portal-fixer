package parser

import (
	"regexp"
	"strconv"
	"strings"
)

var salaryRangeRE = regexp.MustCompile(`\$([0-9][0-9,]*)\s*(?:-|to|–)\s*\$?([0-9][0-9,]*)`)

func SalaryRange(text string) (int, int, bool) {
	match := salaryRangeRE.FindStringSubmatch(text)
	if len(match) != 3 {
		return 0, 0, false
	}

	min, err := strconv.Atoi(strings.ReplaceAll(match[1], ",", ""))
	if err != nil {
		return 0, 0, false
	}

	max, err := strconv.Atoi(strings.ReplaceAll(match[2], ",", ""))
	if err != nil {
		return 0, 0, false
	}

	return min, max, true
}

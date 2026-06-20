package web

import "strconv"

func formatMoney(v int) string {
	raw := strconv.Itoa(v)
	out := ""
	for i, r := range reverse(raw) {
		if i > 0 && i%3 == 0 {
			out = "," + out
		}
		out = string(r) + out
	}
	return "$" + out
}

func reverse(input string) []rune {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return runes
}

package search_engine

func Frequencies(tokens []string) map[string]int {
	out := make(map[string]int)
	for _, token := range tokens {
		out[token]++
	}
	return out
}

package search_engine

var stopWords = map[string]bool{
	"and":  true,
	"are":  true,
	"for":  true,
	"from": true,
	"has":  true,
	"have": true,
	"the":  true,
	"this": true,
	"that": true,
	"with": true,
	"will": true,
	"you":  true,
	"your": true,
}

func isStopWord(term string) bool {
	return stopWords[term]
}

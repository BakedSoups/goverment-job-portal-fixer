package search_engine

import (
	"sort"
	"strings"
)

type Result struct {
	Document Document
	Score    int
	Reasons  []string
}

func (e *Engine) Search(query string) []Result {
	query = strings.TrimSpace(query)
	if query == "" {
		docs := e.Documents()
		results := make([]Result, 0, len(docs))
		for _, doc := range docs {
			results = append(results, Result{Document: doc, Score: 1})
		}
		sortResults(results)
		return results
	}

	terms := e.CanonicalTerms(query)
	results := make([]Result, 0)

	for _, doc := range e.Documents() {
		score, reasons := e.score(doc, terms)
		if score > 0 {
			results = append(results, Result{Document: doc, Score: score, Reasons: reasons})
		}
	}

	sortResults(results)
	return results
}

func (e *Engine) score(doc Document, terms []string) (int, []string) {
	score := 0
	var reasons []string
	title := strings.ToLower(doc.Job.Title)
	department := strings.ToLower(doc.Job.Department)

	for _, term := range terms {
		if count := doc.ConceptHits[term]; count > 0 {
			points := count * 12
			score += points
			reasons = append(reasons, term+" concept matched")
		}

		if count := doc.Frequencies[term]; count > 0 {
			points := count * 3
			score += points
			reasons = append(reasons, term+" frequency matched")
		}

		if strings.Contains(title, strings.ReplaceAll(term, "_", " ")) || strings.Contains(title, term) {
			score += 30
			reasons = append(reasons, term+" matched title")
		}

		if strings.Contains(department, term) {
			score += 8
			reasons = append(reasons, term+" matched department")
		}
	}

	return score, unique(reasons)
}

func sortResults(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Document.Job.ReleasedDate.After(results[j].Document.Job.ReleasedDate)
		}
		return results[i].Score > results[j].Score
	})
}

func unique(input []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range input {
		if !seen[item] {
			out = append(out, item)
			seen[item] = true
		}
	}
	return out
}

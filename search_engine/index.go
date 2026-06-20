package search_engine

import (
	"sort"
	"strings"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/taxonomy"
)

type Engine struct {
	jobs       []jobs.Job
	docs       map[string]Document
	matcher    Matcher
	aliasMap   map[string]taxonomy.Concept
	concepts   map[string]taxonomy.Concept
	globalFreq map[string]int
}

type Document struct {
	Job          jobs.Job
	Frequencies  map[string]int
	ConceptHits  map[string]int
	ConceptNames []string
}

func NewEngine(input []jobs.Job) *Engine {
	matcher := NewMatcher()
	engine := &Engine{
		jobs:       input,
		docs:       make(map[string]Document),
		matcher:    matcher,
		aliasMap:   taxonomy.AliasMap(),
		concepts:   make(map[string]taxonomy.Concept),
		globalFreq: make(map[string]int),
	}

	for _, concept := range taxonomy.Concepts() {
		engine.concepts[concept.Name] = concept
	}

	for _, job := range input {
		tokens := Tokens(job.FullText)
		frequencies := Frequencies(tokens)
		concepts := matcher.Match(job.FullText)
		names := make([]string, 0, len(concepts))
		for name, count := range concepts {
			if count <= 0 {
				continue
			}
			names = append(names, name)
			engine.globalFreq[name] += count
		}
		sort.Strings(names)

		for token, count := range frequencies {
			engine.globalFreq[token] += count
		}

		engine.docs[job.ID] = Document{
			Job:          job,
			Frequencies:  frequencies,
			ConceptHits:  concepts,
			ConceptNames: names,
		}
	}

	return engine
}

func (e *Engine) Documents() []Document {
	out := make([]Document, 0, len(e.jobs))
	for _, job := range e.jobs {
		out = append(out, e.docs[job.ID])
	}
	return out
}

func (e *Engine) Document(id string) (Document, bool) {
	doc, ok := e.docs[id]
	return doc, ok
}

func (e *Engine) TopTerms(limit int) []TermCount {
	terms := make([]TermCount, 0, len(e.globalFreq))
	for term, count := range e.globalFreq {
		concept, ok := e.concepts[term]
		if !ok || len(term) < 3 || isStopWord(term) {
			continue
		}
		terms = append(terms, TermCount{Term: term, Label: concept.Label, Count: count})
	}
	sort.Slice(terms, func(i, j int) bool {
		if terms[i].Count == terms[j].Count {
			return terms[i].Term < terms[j].Term
		}
		return terms[i].Count > terms[j].Count
	})
	if len(terms) > limit {
		return terms[:limit]
	}
	return terms
}

func (e *Engine) CanonicalTerms(query string) []string {
	var out []string
	seen := map[string]bool{}
	lower := strings.ToLower(query)

	for alias, concept := range e.aliasMap {
		if strings.Contains(lower, alias) && !seen[concept.Name] {
			out = append(out, concept.Name)
			seen[concept.Name] = true
		}
	}

	for _, token := range Tokens(query) {
		if concept, ok := e.aliasMap[token]; ok {
			token = concept.Name
		}
		if !seen[token] {
			out = append(out, token)
			seen[token] = true
		}
	}

	sort.Strings(out)
	return out
}

type TermCount struct {
	Term  string
	Label string
	Count int
}

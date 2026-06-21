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

type Tag struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Category string   `json:"category"`
	Aliases  []string `json:"aliases"`
}

type TagEvidence struct {
	Tag      Tag
	Snippets []string
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
	terms := make([]TermCount, 0, len(e.concepts))
	for term, concept := range e.concepts {
		if len(term) < 3 || isStopWord(term) {
			continue
		}
		count := 0
		for _, doc := range e.docs {
			if doc.ConceptHits[term] > 0 {
				count++
			}
		}
		if count == 0 {
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

func (e *Engine) Tags() []Tag {
	tags := make([]Tag, 0, len(e.concepts))
	for _, concept := range e.concepts {
		tags = append(tags, Tag{
			ID:       concept.Name,
			Label:    concept.Label,
			Category: concept.Category,
			Aliases:  e.aliasesForTag(concept.Name),
		})
	}
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Category == tags[j].Category {
			return tags[i].Label < tags[j].Label
		}
		return tags[i].Category < tags[j].Category
	})
	return tags
}

func (e *Engine) Tag(id string) (Tag, bool) {
	concept, ok := e.concepts[id]
	if !ok {
		return Tag{}, false
	}
	return Tag{ID: concept.Name, Label: concept.Label, Category: concept.Category, Aliases: e.aliasesForTag(concept.Name)}, true
}

func (e *Engine) Evidence(doc Document, tagIDs []string, limitPerTag int) []TagEvidence {
	out := make([]TagEvidence, 0, len(tagIDs))
	for _, id := range tagIDs {
		tag, ok := e.Tag(id)
		if !ok {
			continue
		}
		aliases := e.aliasesForTag(id)
		snippets := evidenceSnippets(doc.Job.Sections, aliases, limitPerTag)
		if len(snippets) == 0 {
			continue
		}
		out = append(out, TagEvidence{Tag: tag, Snippets: snippets})
	}
	return out
}

func (e *Engine) aliasesForTag(id string) []string {
	concept, ok := e.concepts[id]
	if !ok {
		return nil
	}
	aliases := append([]string{concept.Name, concept.Label}, concept.Aliases...)
	for i := range aliases {
		aliases[i] = strings.ToLower(strings.ReplaceAll(aliases[i], "_", " "))
	}
	return aliases
}

func evidenceSnippets(sections []jobs.Section, aliases []string, limit int) []string {
	var snippets []string
	seen := map[string]bool{}
	for _, section := range sections {
		lines := strings.Split(section.Text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || seen[line] {
				continue
			}
			lower := strings.ToLower(line)
			for _, alias := range aliases {
				if alias != "" && strings.Contains(lower, alias) {
					snippets = append(snippets, line)
					seen[line] = true
					break
				}
			}
			if limit > 0 && len(snippets) >= limit {
				return snippets
			}
		}
	}
	return snippets
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

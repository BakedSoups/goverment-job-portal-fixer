package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
)

type IndexPage struct {
	PageTitle       string
	Query           string
	SelectedTags    []search_engine.Tag
	AvailableTags   []search_engine.Tag
	SelectedTagIDs  string
	YOEActive       bool
	YOE             int
	HiddenByYOE     int
	Count           int
	UnfilteredCount int
	Results         []ResultView
	TopTerms        []search_engine.TermCount
	SearchParams    string
}

type ResultView struct {
	search_engine.Result
	SearchParams string
}

type JobPage struct {
	PageTitle    string
	Document     search_engine.Document
	SelectedTags []search_engine.Tag
	Evidence     []search_engine.TagEvidence
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	selectedTagIDs := parseTags(r.URL.Query().Get("tags"))
	results := s.engine.SearchTags(selectedTagIDs)
	if len(selectedTagIDs) == 0 {
		results = s.engine.Search(query)
	}
	unfilteredCount := len(results)

	yoe, yoeActive := parseYOE(r.URL.Query().Get("yoe"))
	if yoeActive {
		results = filterByYOE(results, yoe)
	}

	selectedTags := make([]search_engine.Tag, 0, len(selectedTagIDs))
	for _, id := range selectedTagIDs {
		if tag, ok := s.engine.Tag(id); ok {
			selectedTags = append(selectedTags, tag)
		}
	}

	page := IndexPage{
		PageTitle:       "SF Jobs Index",
		Query:           query,
		SelectedTags:    selectedTags,
		AvailableTags:   s.engine.Tags(),
		SelectedTagIDs:  strings.Join(selectedTagIDs, ","),
		YOEActive:       yoeActive,
		YOE:             yoe,
		HiddenByYOE:     unfilteredCount - len(results),
		Count:           len(results),
		UnfilteredCount: unfilteredCount,
		Results:         resultViews(results, searchParams(selectedTagIDs, yoe, yoeActive)),
		TopTerms:        s.engine.TopTerms(18),
		SearchParams:    searchParams(selectedTagIDs, yoe, yoeActive),
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func resultViews(results []search_engine.Result, params string) []ResultView {
	out := make([]ResultView, 0, len(results))
	for _, result := range results {
		out = append(out, ResultView{Result: result, SearchParams: params})
	}
	return out
}

func parseYOE(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	yoe, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if yoe < 0 {
		yoe = 0
	}
	if yoe > 10 {
		yoe = 10
	}
	return yoe, true
}

func filterByYOE(results []search_engine.Result, yoe int) []search_engine.Result {
	filtered := make([]search_engine.Result, 0, len(results))
	for _, result := range results {
		job := result.Document.Job
		if !job.RequiredYOEFound || job.RequiredYOEMin <= yoe {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func parseTags(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		out = append(out, part)
		seen[part] = true
	}
	return out
}

func (s *Server) handleJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	id = strings.TrimSpace(id)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	doc, ok := s.engine.Document(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	selectedTagIDs := parseTags(r.URL.Query().Get("tags"))
	selectedTags := make([]search_engine.Tag, 0, len(selectedTagIDs))
	for _, id := range selectedTagIDs {
		if tag, ok := s.engine.Tag(id); ok {
			selectedTags = append(selectedTags, tag)
		}
	}

	page := JobPage{
		PageTitle:    doc.Job.Title,
		Document:     doc,
		SelectedTags: selectedTags,
		Evidence:     s.engine.Evidence(doc, selectedTagIDs, 3),
	}
	if err := s.templates.ExecuteTemplate(w, "job.html", page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func searchParams(tags []string, yoe int, yoeActive bool) string {
	var parts []string
	if len(tags) > 0 {
		parts = append(parts, "tags="+strings.Join(tags, ","))
	}
	if yoeActive {
		parts = append(parts, "yoe="+strconv.Itoa(yoe))
	}
	if len(parts) == 0 {
		return ""
	}
	return "?" + strings.Join(parts, "&")
}

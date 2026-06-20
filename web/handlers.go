package web

import (
	"net/http"
	"strings"

	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
)

type IndexPage struct {
	PageTitle  string
	Query      string
	Count      int
	Results    []search_engine.Result
	TopTerms   []search_engine.TermCount
	SearchHint string
}

type JobPage struct {
	PageTitle string
	Document  search_engine.Document
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	results := s.engine.Search(query)
	page := IndexPage{
		PageTitle:  "SF Jobs Index",
		Query:      query,
		Count:      len(results),
		Results:    results,
		TopTerms:   s.engine.TopTerms(18),
		SearchHint: "python software engineer, data sql, nosql analyst",
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	page := JobPage{PageTitle: doc.Job.Title, Document: doc}
	if err := s.templates.ExecuteTemplate(w, "job.html", page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
)

type Server struct {
	engine    *search_engine.Engine
	jobs      []jobs.Job
	templates *template.Template
}

func NewServer(engine *search_engine.Engine, input []jobs.Job) (*Server, error) {
	funcs := template.FuncMap{
		"date": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("Jan 2, 2006")
		},
		"money": func(v int) string {
			if v == 0 {
				return ""
			}
			return formatMoney(v)
		},
		"plural": func(n int, singular string, plural string) string {
			if n == 1 {
				return singular
			}
			return plural
		},
		"lines": func(v string) []string {
			return strings.Split(v, "\n")
		},
		"json": func(v any) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			return template.JS(b)
		},
	}

	patterns := []string{
		filepath.Join("templates", "*.html"),
		filepath.Join("templates", "partials", "*.html"),
	}

	tmpl, err := template.New("").Funcs(funcs).ParseGlob(patterns[0])
	if err != nil {
		return nil, err
	}
	tmpl, err = tmpl.ParseGlob(patterns[1])
	if err != nil {
		return nil, err
	}

	return &Server{engine: engine, jobs: input, templates: tmpl}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/jobs/", s.handleJob)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

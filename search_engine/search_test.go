package search_engine

import (
	"testing"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
)

func TestSearchMatchesConceptAliases(t *testing.T) {
	engine := NewEngine([]jobs.Job{
		{
			ID:       "1",
			Title:    "Application Developer",
			FullText: "Build services with Python, SQL, and reporting dashboards.",
		},
		{
			ID:       "2",
			Title:    "Clerk",
			FullText: "Process applications and answer phones.",
		},
	})

	results := engine.Search("py sql")
	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}
	if results[0].Document.Job.ID != "1" {
		t.Fatalf("Search() top job = %q, want 1", results[0].Document.Job.ID)
	}
}

func TestSearchRequiresEveryQueryTerm(t *testing.T) {
	engine := NewEngine([]jobs.Job{
		{
			ID:       "1",
			Title:    "Application Developer",
			FullText: "Build services with Python and SQL.",
		},
		{
			ID:       "2",
			Title:    "Probation Officer",
			FullText: "Coordinate probation case work.",
		},
	})

	results := engine.Search("python + probation")
	if len(results) != 0 {
		t.Fatalf("Search() returned %d results, want 0", len(results))
	}
}

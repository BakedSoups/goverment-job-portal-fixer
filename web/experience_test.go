package web

import (
	"testing"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
)

func TestFilterByYOERange(t *testing.T) {
	results := []search_engine.Result{
		{Document: search_engine.Document{Job: jobs.Job{ID: "junior", RequiredYOEFound: true, RequiredYOEMin: 1, RequiredYOEMax: 1}}},
		{Document: search_engine.Document{Job: jobs.Job{ID: "mid", RequiredYOEFound: true, RequiredYOEMin: 4, RequiredYOEMax: 4}}},
		{Document: search_engine.Document{Job: jobs.Job{ID: "senior", RequiredYOEFound: true, RequiredYOEMin: 8, RequiredYOEMax: 8}}},
		{Document: search_engine.Document{Job: jobs.Job{ID: "unspecified"}}},
	}

	filtered := filterByYOERange(results, 2, 6)
	if len(filtered) != 1 || filtered[0].Document.Job.ID != "mid" {
		t.Fatalf("2-6 year filter returned %+v, want only mid", filtered)
	}
}

func TestParseYOERange(t *testing.T) {
	minYOE, maxYOE := parseYOERange("6", "2", "")
	if minYOE != 2 || maxYOE != 6 {
		t.Fatalf("reversed range = %d-%d, want 2-6", minYOE, maxYOE)
	}

	minYOE, maxYOE = parseYOERange("", "", "4")
	if minYOE != 0 || maxYOE != 4 {
		t.Fatalf("legacy maximum = %d-%d, want 0-4", minYOE, maxYOE)
	}
}

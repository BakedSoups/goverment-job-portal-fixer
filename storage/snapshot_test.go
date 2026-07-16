package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJobs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jobs.json")
	contents := `{"generatedAt":"2026-07-16T00:00:00Z","documents":[{"Job":{"ID":"job-1","Title":"Engineer","SourceName":"Oakland"},"ConceptHits":{"go":1}}],"tags":[]}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadJobs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].ID != "job-1" || loaded[0].Title != "Engineer" {
		t.Fatalf("unexpected jobs: %#v", loaded)
	}
}

func TestLoadJobsRejectsEmptySnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jobs.json")
	if err := os.WriteFile(path, []byte(`{"documents":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadJobs(path); err == nil {
		t.Fatal("LoadJobs() error = nil, want empty snapshot error")
	}
}

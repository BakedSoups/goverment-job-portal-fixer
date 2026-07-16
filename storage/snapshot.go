package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
)

type snapshot struct {
	Documents []snapshotDocument `json:"documents"`
}

type snapshotDocument struct {
	Job jobs.Job `json:"Job"`
}

func LoadJobs(path string) ([]jobs.Job, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open snapshot %s: %w", path, err)
	}
	defer file.Close()

	var cached snapshot
	if err := json.NewDecoder(file).Decode(&cached); err != nil {
		return nil, fmt.Errorf("decode snapshot %s: %w", path, err)
	}
	if len(cached.Documents) == 0 {
		return nil, fmt.Errorf("snapshot %s contains no jobs", path)
	}

	loaded := make([]jobs.Job, 0, len(cached.Documents))
	for _, document := range cached.Documents {
		loaded = append(loaded, document.Job)
	}
	return loaded, nil
}

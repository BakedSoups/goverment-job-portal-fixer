package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/scraper"
	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
)

type snapshot struct {
	GeneratedAt time.Time                `json:"generatedAt"`
	Documents   []search_engine.Document `json:"documents"`
	Tags        []search_engine.Tag      `json:"tags"`
}

func main() {
	output := flag.String("output", "", "path for the generated JSON snapshot")
	flag.Parse()
	if *output == "" {
		fmt.Fprintln(os.Stderr, "export-snapshot: -output is required")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	rawJobs, err := scraper.NewClient(http.DefaultClient).FetchAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "export-snapshot: fetch jobs: %v\n", err)
		os.Exit(1)
	}

	normalized := make([]jobs.Job, 0, len(rawJobs))
	for _, raw := range rawJobs {
		normalized = append(normalized, jobs.FromSmartRecruiters(raw))
	}
	engine := search_engine.NewEngine(normalized)

	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "export-snapshot: create output directory: %v\n", err)
		os.Exit(1)
	}
	file, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "export-snapshot: create output: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(snapshot{GeneratedAt: time.Now().UTC(), Documents: engine.Documents(), Tags: engine.Tags()}); err != nil {
		fmt.Fprintf(os.Stderr, "export-snapshot: encode snapshot: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("exported %d jobs to %s\n", len(normalized), *output)
}

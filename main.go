package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/scraper"
	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
	"github.com/BakedSoups/goverment-job-portal-fixer/storage"
	"github.com/BakedSoups/goverment-job-portal-fixer/web"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	normalized, err := loadJobs(ctx)
	if err != nil {
		log.Fatalf("load jobs: %v", err)
	}

	engine := search_engine.NewEngine(normalized)
	server, err := web.NewServer(engine, normalized)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	addr := listenAddress()

	log.Printf("loaded %d jobs", len(normalized))
	log.Printf("listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

func loadJobs(ctx context.Context) ([]jobs.Job, error) {
	if snapshotPath := strings.TrimSpace(os.Getenv("JOBS_SNAPSHOT")); snapshotPath != "" {
		return storage.LoadJobs(snapshotPath)
	}

	rawJobs, err := scraper.NewClient(http.DefaultClient).FetchAll(ctx)
	if err != nil {
		return nil, err
	}
	normalized := make([]jobs.Job, 0, len(rawJobs))
	for _, raw := range rawJobs {
		normalized = append(normalized, jobs.FromSmartRecruiters(raw))
	}
	return normalized, nil
}

func listenAddress() string {
	if addr := strings.TrimSpace(os.Getenv("ADDR")); addr != "" {
		return addr
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + port
	}
	return ":8080"
}

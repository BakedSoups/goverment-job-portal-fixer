package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/scraper"
	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
	"github.com/BakedSoups/goverment-job-portal-fixer/web"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client := scraper.NewClient(http.DefaultClient)
	rawJobs, err := client.FetchAll(ctx)
	if err != nil {
		log.Fatalf("scrape jobs: %v", err)
	}

	normalized := make([]jobs.Job, 0, len(rawJobs))
	for _, raw := range rawJobs {
		normalized = append(normalized, jobs.FromSmartRecruiters(raw))
	}

	engine := search_engine.NewEngine(normalized)
	server, err := web.NewServer(engine, normalized)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("loaded %d jobs", len(normalized))
	log.Printf("listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

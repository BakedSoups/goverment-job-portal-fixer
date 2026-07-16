package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
	"github.com/BakedSoups/goverment-job-portal-fixer/web"
)

func TestMapRendersWithGeoJSONAsset(t *testing.T) {
	input := []jobs.Job{{
		ID:           "east-bay-job",
		SourceID:     "oakland",
		SourceName:   "Oakland",
		SourceRegion: "East Bay",
		Title:        "Analyst",
	}}
	server, err := web.NewServer(search_engine.NewEngine(input), input)
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	handler := server.Routes()

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/?gov=oakland&q=Analyst", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("index status = %d, want %d", page.Code, http.StatusOK)
	}
	if !strings.Contains(page.Body.String(), `data-regions-url="/static/data/bay-area-regions.geojson"`) {
		t.Fatal("index does not reference the Bay Area GeoJSON asset")
	}
	if strings.Contains(page.Body.String(), "data-sources=") {
		t.Fatal("index still includes the removed map marker payload")
	}
	if !strings.Contains(page.Body.String(), `class="job-card" data-region="East Bay"`) {
		t.Fatal("job card does not expose its region for matching map colors")
	}
	if !strings.Contains(page.Body.String(), "Required experience: not specified") {
		t.Fatal("job card does not use the user-facing unspecified experience fallback")
	}
	if strings.Contains(page.Body.String(), "Showing all current listings.") || strings.Contains(page.Body.String(), "Government filter active.") {
		t.Fatal("results header still renders verbose filter context")
	}
	if !strings.Contains(page.Body.String(), `class="match-label" title="Your search terms appear in the job title.">Title match`) {
		t.Fatal("job card does not render a plain-language match label")
	}
	resultsAt := strings.Index(page.Body.String(), `class="results-column"`)
	filtersAt := strings.Index(page.Body.String(), `class="filter-sidebar"`)
	if resultsAt < 0 || filtersAt < 0 || resultsAt > filtersAt {
		t.Fatal("index does not render results before the filter sidebar")
	}
	if strings.Contains(page.Body.String(), `<select id="government-select"`) {
		t.Fatal("index still renders the old single-select government filter")
	}
	if !strings.Contains(page.Body.String(), `type="checkbox" name="gov" value="oakland" checked`) {
		t.Fatal("index does not render the selected government checkbox")
	}
	if !strings.Contains(page.Body.String(), `data-government-summary>1 selected`) {
		t.Fatal("government picker does not summarize the selected count")
	}
	if !strings.Contains(page.Body.String(), `data-region="East Bay" open`) {
		t.Fatal("selected region dropdown is not expanded")
	}
	if !strings.Contains(page.Body.String(), `data-region-toggle aria-label="Select all East Bay governments" checked`) {
		t.Fatal("region dropdown does not render a selected region toggle")
	}
	if strings.Contains(page.Body.String(), ">Apply filters</button>") {
		t.Fatal("government picker still requires an Apply filters button")
	}
	if !strings.Contains(page.Body.String(), `data-live-status aria-live="polite"`) {
		t.Fatal("filter form does not include an accessible live-update status")
	}
	if !strings.Contains(page.Body.String(), "Job information provided through public APIs") {
		t.Fatal("page header does not disclose the public API data source")
	}
	if !strings.Contains(page.Body.String(), `href="https://github.com/BakedSoups/goverment-job-portal-fixer"`) {
		t.Fatal("page header does not link to the GitHub repository")
	}
	if !strings.Contains(page.Body.String(), `data-theme-toggle aria-pressed="false"`) {
		t.Fatal("page header does not render the dark-mode control")
	}
	if !strings.Contains(page.Body.String(), `class="yoe-range-track" data-yoe-track`) ||
		!strings.Contains(page.Body.String(), `name="yoe_min"`) ||
		!strings.Contains(page.Body.String(), `name="yoe_max"`) {
		t.Fatal("experience filter does not render two handles on one range track")
	}

	asset := httptest.NewRecorder()
	handler.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/static/data/bay-area-regions.geojson", nil))
	if asset.Code != http.StatusOK {
		t.Fatalf("GeoJSON status = %d, want %d", asset.Code, http.StatusOK)
	}
	if !strings.Contains(asset.Body.String(), `"type":"FeatureCollection"`) {
		t.Fatal("GeoJSON response is not a feature collection")
	}

	detail := httptest.NewRecorder()
	handler.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/jobs/east-bay-job", nil))
	if detail.Code != http.StatusOK {
		t.Fatalf("job detail status = %d, want %d", detail.Code, http.StatusOK)
	}
	if !strings.Contains(detail.Body.String(), "Click a signal to find where it was found in the listing.") {
		t.Fatal("job detail does not explain parsed-signal interaction")
	}
	if !strings.Contains(detail.Body.String(), "<strong>Required experience:</strong> not specified") {
		t.Fatal("job detail does not use the user-facing unspecified experience fallback")
	}
}

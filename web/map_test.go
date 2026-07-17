package web

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/BakedSoups/goverment-job-portal-fixer/jobs"
	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
)

func TestMapDistrictsMatchGeoJSONRegions(t *testing.T) {
	data, err := os.ReadFile("../static/data/bay-area-regions.geojson")
	if err != nil {
		t.Fatalf("read map regions: %v", err)
	}

	var collection struct {
		Type     string `json:"type"`
		Features []struct {
			Properties struct {
				ID       string   `json:"id"`
				Counties []string `json:"counties"`
			} `json:"properties"`
			Geometry struct {
				Type string `json:"type"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("parse map regions: %v", err)
	}
	if collection.Type != "FeatureCollection" {
		t.Fatalf("collection type = %q, want FeatureCollection", collection.Type)
	}

	results := []search_engine.Result{
		{Document: search_engine.Document{Job: jobs.Job{SourceRegion: "East Bay"}}},
		{Document: search_engine.Document{Job: jobs.Job{SourceRegion: "East Bay"}}},
		{Document: search_engine.Document{Job: jobs.Job{SourceRegion: "SF"}}},
	}
	districts := mapDistricts(results, []string{"oakland"})
	if len(districts) != len(collection.Features) {
		t.Fatalf("got %d district records for %d GeoJSON features", len(districts), len(collection.Features))
	}

	byID := make(map[string]MapDistrict, len(districts))
	for _, district := range districts {
		byID[district.ID] = district
	}
	for _, feature := range collection.Features {
		if feature.Geometry.Type != "Polygon" && feature.Geometry.Type != "MultiPolygon" {
			t.Errorf("region %q has unsupported geometry %q", feature.Properties.ID, feature.Geometry.Type)
		}
		if len(feature.Properties.Counties) > 1 && feature.Geometry.Type != "Polygon" {
			t.Errorf("multi-county region %q was not dissolved: geometry = %q", feature.Properties.ID, feature.Geometry.Type)
		}
		if _, ok := byID[feature.Properties.ID]; !ok {
			t.Errorf("GeoJSON region %q has no district metadata", feature.Properties.ID)
		}
	}

	eastBay := byID["east-bay"]
	if eastBay.Count != 2 || !eastBay.Selected || eastBay.Level != 4 {
		t.Errorf("East Bay metadata = %+v", eastBay)
	}
}

func TestMapPointsAggregateMatchingJobsBySource(t *testing.T) {
	results := []search_engine.Result{
		{Document: search_engine.Document{Job: jobs.Job{SourceID: "oakland", SourceName: "Oakland", SourceRegion: "East Bay"}}},
		{Document: search_engine.Document{Job: jobs.Job{SourceID: "oakland", SourceName: "Oakland", SourceRegion: "East Bay"}}},
		{Document: search_engine.Document{Job: jobs.Job{SourceID: "sf", SourceName: "San Francisco", SourceRegion: "SF"}}},
		{Document: search_engine.Document{Job: jobs.Job{SourceID: "unknown", SourceName: "Unknown"}}},
	}

	points := mapPoints(results)
	if len(points) != 2 {
		t.Fatalf("got %d map points, want 2", len(points))
	}
	if points[0].Name != "Oakland" || points[0].Count != 2 {
		t.Errorf("Oakland point = %+v", points[0])
	}
	if points[1].Name != "San Francisco" || points[1].Count != 1 {
		t.Errorf("San Francisco point = %+v", points[1])
	}
	if points[0].Latitude == 0 || points[0].Longitude == 0 {
		t.Errorf("Oakland point lacks coordinates: %+v", points[0])
	}
}

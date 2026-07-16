package jobs

import (
	"strings"
	"time"

	"github.com/BakedSoups/goverment-job-portal-fixer/parser"
	"github.com/BakedSoups/goverment-job-portal-fixer/scraper"
)

type Job struct {
	ID                     string
	SourceID               string
	SourceName             string
	SourceRegion           string
	Title                  string
	RefNumber              string
	Department             string
	Location               string
	Employment             string
	Experience             string
	ReleasedDate           time.Time
	PostingURL             string
	ApplyURL               string
	SalaryMin              int
	SalaryMax              int
	RequiredYOEMin         int
	RequiredYOEMax         int
	RequiredYOEFound       bool
	RequiredYOESource      string
	RequiredYOEConfidence  string
	PreferredYOEMin        int
	PreferredYOEMax        int
	PreferredYOEFound      bool
	PreferredYOESource     string
	PreferredYOEConfidence string
	Sections               []Section
	FullText               string
	Fields                 map[string]string
}

type Section struct {
	Key   string
	Title string
	Text  string
}

func FromSmartRecruiters(raw scraper.Posting) Job {
	sections := []Section{
		cleanSection("company", raw.JobAd.Sections.CompanyDescription),
		cleanSection("description", raw.JobAd.Sections.JobDescription),
		cleanSection("qualifications", raw.JobAd.Sections.Qualifications),
		cleanSection("additional", raw.JobAd.Sections.AdditionalInformation),
	}

	parts := []string{raw.Name, raw.RefNumber, raw.Department.Label, raw.Location.FullLocation}
	for _, section := range sections {
		parts = append(parts, section.Title, section.Text)
	}

	fullText := strings.Join(parts, "\n")
	minSalary, maxSalary, _ := parser.SalaryRange(fullText)
	requiredExperience := parser.RequiredExperience(sections[2].Text)
	preferredExperience := parser.PreferredExperience(sections[2].Text)

	released, _ := time.Parse(time.RFC3339Nano, raw.ReleasedDate)

	fields := make(map[string]string, len(raw.CustomField))
	for _, field := range raw.CustomField {
		if field.FieldLabel != "" && field.ValueLabel != "" {
			fields[field.FieldLabel] = field.ValueLabel
		}
	}

	return Job{
		ID:                     raw.ID,
		SourceID:               raw.Source.ID,
		SourceName:             raw.Source.Name,
		SourceRegion:           raw.Source.Region,
		Title:                  raw.Name,
		RefNumber:              raw.RefNumber,
		Department:             raw.Department.Label,
		Location:               raw.Location.FullLocation,
		Employment:             raw.Employment.Label,
		Experience:             raw.Experience.Label,
		ReleasedDate:           released,
		PostingURL:             raw.PostingURL,
		ApplyURL:               raw.ApplyURL,
		SalaryMin:              minSalary,
		SalaryMax:              maxSalary,
		RequiredYOEMin:         requiredExperience.Min,
		RequiredYOEMax:         requiredExperience.Max,
		RequiredYOEFound:       requiredExperience.Found,
		RequiredYOESource:      requiredExperience.Source,
		RequiredYOEConfidence:  requiredExperience.Confidence,
		PreferredYOEMin:        preferredExperience.Min,
		PreferredYOEMax:        preferredExperience.Max,
		PreferredYOEFound:      preferredExperience.Found,
		PreferredYOESource:     preferredExperience.Source,
		PreferredYOEConfidence: preferredExperience.Confidence,
		Sections:               sections,
		FullText:               parser.CleanText(fullText),
		Fields:                 fields,
	}
}

func cleanSection(key string, raw scraper.Section) Section {
	return Section{
		Key:   key,
		Title: raw.Title,
		Text:  parser.HTMLToText(raw.Text),
	}
}

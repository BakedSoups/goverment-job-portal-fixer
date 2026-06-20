package jobs

import (
	"strings"
	"time"

	"github.com/BakedSoups/goverment-job-portal-fixer/parser"
	"github.com/BakedSoups/goverment-job-portal-fixer/scraper"
)

type Job struct {
	ID           string
	Title        string
	RefNumber    string
	Department   string
	Location     string
	Employment   string
	Experience   string
	ReleasedDate time.Time
	PostingURL   string
	ApplyURL     string
	SalaryMin    int
	SalaryMax    int
	Sections     []Section
	FullText     string
	Fields       map[string]string
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

	released, _ := time.Parse(time.RFC3339Nano, raw.ReleasedDate)

	fields := make(map[string]string, len(raw.CustomField))
	for _, field := range raw.CustomField {
		if field.FieldLabel != "" && field.ValueLabel != "" {
			fields[field.FieldLabel] = field.ValueLabel
		}
	}

	return Job{
		ID:           raw.ID,
		Title:        raw.Name,
		RefNumber:    raw.RefNumber,
		Department:   raw.Department.Label,
		Location:     raw.Location.FullLocation,
		Employment:   raw.Employment.Label,
		Experience:   raw.Experience.Label,
		ReleasedDate: released,
		PostingURL:   raw.PostingURL,
		ApplyURL:     raw.ApplyURL,
		SalaryMin:    minSalary,
		SalaryMax:    maxSalary,
		Sections:     sections,
		FullText:     parser.CleanText(fullText),
		Fields:       fields,
	}
}

func cleanSection(key string, raw scraper.Section) Section {
	return Section{
		Key:   key,
		Title: raw.Title,
		Text:  parser.HTMLToText(raw.Text),
	}
}

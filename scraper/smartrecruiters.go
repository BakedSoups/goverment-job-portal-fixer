package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	smartRecruitersAPIBase = "https://api.smartrecruiters.com/v1/companies"
	governmentJobsBase     = "https://www.governmentjobs.com/careers"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient}
}

func Sources() []Source {
	return []Source{
		{ID: "sf", Name: "San Francisco", Kind: "smartrecruiters", Identifier: "CityAndCountyOfSanFrancisco1", Region: "SF"},
		{ID: "contra-costa-county", Name: "Contra Costa County", Kind: "governmentjobs", Identifier: "contracosta", Region: "East Bay"},
		{ID: "san-mateo-county", Name: "San Mateo County", Kind: "governmentjobs", Identifier: "sanmateo", Region: "Peninsula"},
		{ID: "santa-clara-county", Name: "Santa Clara County", Kind: "governmentjobs", Identifier: "santaclara", Region: "South Bay"},
		{ID: "marin-county", Name: "Marin County", Kind: "governmentjobs", Identifier: "marin", Region: "North Bay"},
		{ID: "sonoma-county", Name: "Sonoma County", Kind: "governmentjobs", Identifier: "sonoma", Region: "North Bay"},
		{ID: "napa-county", Name: "Napa County", Kind: "governmentjobs", Identifier: "napacounty", Region: "North Bay"},
		{ID: "solano-county", Name: "Solano County", Kind: "governmentjobs", Identifier: "solanocounty", Region: "North Bay"},
		{ID: "alameda", Name: "City of Alameda", Kind: "governmentjobs", Identifier: "alamedaca", Region: "East Bay"},
		{ID: "berkeley", Name: "Berkeley", Kind: "governmentjobs", Identifier: "berkeley", Region: "East Bay"},
		{ID: "fremont", Name: "Fremont", Kind: "governmentjobs", Identifier: "fremontca", Region: "East Bay"},
		{ID: "hayward", Name: "Hayward", Kind: "governmentjobs", Identifier: "haywardca", Region: "East Bay"},
		{ID: "oakland", Name: "Oakland", Kind: "governmentjobs", Identifier: "oaklandca", Region: "East Bay"},
		{ID: "richmond", Name: "Richmond", Kind: "governmentjobs", Identifier: "richmondca", Region: "East Bay"},
		{ID: "mountain-view", Name: "Mountain View", Kind: "governmentjobs", Identifier: "mountainview", Region: "South Bay"},
		{ID: "palo-alto", Name: "Palo Alto", Kind: "governmentjobs", Identifier: "paloaltoca", Region: "South Bay"},
		{ID: "san-jose", Name: "San Jose", Kind: "governmentjobs", Identifier: "sanjoseca", Region: "South Bay"},
		{ID: "santa-clara", Name: "Santa Clara", Kind: "governmentjobs", Identifier: "cityofsantaclaraca", Region: "South Bay"},
		{ID: "sunnyvale", Name: "Sunnyvale", Kind: "governmentjobs", Identifier: "sunnyvale", Region: "South Bay"},
		{ID: "vallejo", Name: "Vallejo", Kind: "governmentjobs", Identifier: "vallejo", Region: "North Bay"},
	}
}

func (c *Client) FetchAll(ctx context.Context) ([]Posting, error) {
	var postings []Posting
	var failures []string
	for _, source := range Sources() {
		sourcePostings, err := c.fetchSource(ctx, source)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", source.Name, err))
			continue
		}
		postings = append(postings, sourcePostings...)
	}
	if len(postings) == 0 && len(failures) > 0 {
		return nil, fmt.Errorf("fetch sources: %s", strings.Join(failures, "; "))
	}
	return postings, nil
}

func (c *Client) fetchSource(ctx context.Context, source Source) ([]Posting, error) {
	switch source.Kind {
	case "smartrecruiters":
		return c.fetchSmartRecruiters(ctx, source)
	case "governmentjobs":
		return c.fetchGovernmentJobs(ctx, source)
	default:
		return nil, fmt.Errorf("unknown source kind %q", source.Kind)
	}
}

func (c *Client) fetchSmartRecruiters(ctx context.Context, source Source) ([]Posting, error) {
	const limit = 100

	var postings []Posting
	for offset := 0; ; offset += limit {
		listURL := fmt.Sprintf("%s/%s/postings?limit=%d&offset=%d", smartRecruitersAPIBase, source.Identifier, limit, offset)
		list, err := getJSON[ListResponse](ctx, c.httpClient, listURL)
		if err != nil {
			return nil, err
		}

		for _, summary := range list.Content {
			posting, err := getJSON[Posting](ctx, c.httpClient, summary.Ref)
			if err != nil {
				return nil, fmt.Errorf("fetch posting %s: %w", summary.ID, err)
			}
			posting.Source = source
			postings = append(postings, posting)
			time.Sleep(75 * time.Millisecond)
		}

		if offset+limit >= list.TotalFound || len(list.Content) == 0 {
			break
		}
	}

	return postings, nil
}

type governmentJobsResponse struct {
	Success bool                    `json:"success"`
	JobList []governmentJobsPosting `json:"jobList"`
}

type governmentJobsPosting struct {
	ID              int      `json:"ID"`
	Classification  string   `json:"Classification"`
	Location        string   `json:"Location"`
	JobType         string   `json:"JobType"`
	SalaryInfo      string   `json:"SalaryInfo"`
	FullDescription string   `json:"FullDescription"`
	OpenDate        string   `json:"OpenDate"`
	CloseDate       string   `json:"CloseDate"`
	PostingDate     string   `json:"PostingDate"`
	ClosingDate     string   `json:"ClosingDate"`
	DepartmentName  string   `json:"DepartmentName"`
	JobNumber       string   `json:"JobNumber"`
	ExamType        string   `json:"ExamType"`
	Categories      []string `json:"Categories"`
}

func (c *Client) fetchGovernmentJobs(ctx context.Context, source Source) ([]Posting, error) {
	listURL := fmt.Sprintf("%s/%s/home/loadJobsOnMaps", governmentJobsBase, source.Identifier)
	list, err := getGovernmentJobsJSON(ctx, c.httpClient, listURL)
	if err != nil {
		return nil, err
	}
	if !list.Success {
		return nil, fmt.Errorf("load jobs response was not successful")
	}

	postings := make([]Posting, 0, len(list.JobList))
	for _, job := range list.JobList {
		postings = append(postings, governmentJobsToPosting(source, job))
	}
	return postings, nil
}

func governmentJobsToPosting(source Source, job governmentJobsPosting) Posting {
	title := strings.TrimSpace(job.Classification)
	id := fmt.Sprintf("%s-%d", source.ID, job.ID)
	postingURL := fmt.Sprintf("%s/%s/jobs/%d", governmentJobsBase, source.Identifier, job.ID)
	description := strings.TrimSpace(job.FullDescription)
	if job.SalaryInfo != "" {
		description = "Salary: " + job.SalaryInfo + "\n" + description
	}
	if job.ClosingDate != "" || job.CloseDate != "" {
		description += "\nClose Date: " + firstNonEmpty(job.ClosingDate, job.CloseDate)
	}
	if len(job.Categories) > 0 {
		description += "\nCategories: " + strings.Join(job.Categories, ", ")
	}

	return Posting{
		ID:           id,
		Name:         title,
		RefNumber:    job.JobNumber,
		ReleasedDate: parseGovernmentJobsDate(job.PostingDate),
		PostingURL:   postingURL,
		ApplyURL:     postingURL,
		Location:     Location{FullLocation: job.Location},
		Department:   Label{Label: job.DepartmentName},
		Employment:   Label{Label: job.JobType},
		Experience:   Label{Label: job.ExamType},
		CustomField: []CustomField{
			{FieldLabel: "Source", ValueLabel: source.Name},
			{FieldLabel: "Close Date", ValueLabel: firstNonEmpty(job.ClosingDate, job.CloseDate)},
			{FieldLabel: "Salary", ValueLabel: job.SalaryInfo},
		},
		JobAd: JobAd{Sections: Sections{
			CompanyDescription: Section{Title: "Government", Text: source.Name},
			JobDescription:     Section{Title: "Description", Text: description},
			Qualifications:     Section{Title: "Requirements", Text: description},
		}},
		Source: source,
	}
}

func parseGovernmentJobsDate(raw string) string {
	t, err := time.Parse("01/02/06", strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func getJSON[T any](ctx context.Context, httpClient *http.Client, url string) (T, error) {
	var out T

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("User-Agent", "sf-jobs-portal-fixer/0.1")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, fmt.Errorf("decode %s: %w", url, err)
	}

	return out, nil
}

func getGovernmentJobsJSON(ctx context.Context, httpClient *http.Client, url string) (governmentJobsResponse, error) {
	var out governmentJobsResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("User-Agent", "sf-jobs-portal-fixer/0.1")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := httpClient.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, fmt.Errorf("decode %s: %w", url, err)
	}

	return out, nil
}

package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	companyIdentifier = "CityAndCountyOfSanFrancisco1"
	apiBase           = "https://api.smartrecruiters.com/v1/companies"
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

func (c *Client) FetchAll(ctx context.Context) ([]Posting, error) {
	const limit = 100

	var postings []Posting
	for offset := 0; ; offset += limit {
		listURL := fmt.Sprintf("%s/%s/postings?limit=%d&offset=%d", apiBase, companyIdentifier, limit, offset)
		list, err := getJSON[ListResponse](ctx, c.httpClient, listURL)
		if err != nil {
			return nil, err
		}

		for _, summary := range list.Content {
			posting, err := getJSON[Posting](ctx, c.httpClient, summary.Ref)
			if err != nil {
				return nil, fmt.Errorf("fetch posting %s: %w", summary.ID, err)
			}
			postings = append(postings, posting)
			time.Sleep(75 * time.Millisecond)
		}

		if offset+limit >= list.TotalFound || len(list.Content) == 0 {
			break
		}
	}

	return postings, nil
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

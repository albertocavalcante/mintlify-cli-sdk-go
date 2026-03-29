package mintlify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// ErrNoSearchConfig is returned when the search client is not configured.
var ErrNoSearchConfig = errors.New("mintlify: search API not configured (set endpoint and API key)")

// SearchClient queries the Mintlify Search API.
type SearchClient struct {
	endpoint string
	apiKey   string
	http     *http.Client
}

// NewSearchClient creates a SearchClient for the given endpoint and API key.
// The endpoint is the base URL of the Mintlify Search API (e.g. "https://search.mintlify.com").
func NewSearchClient(endpoint, apiKey string) *SearchClient {
	return &SearchClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		http:     http.DefaultClient,
	}
}

// WithHTTPClient overrides the default HTTP client.
func (s *SearchClient) WithHTTPClient(c *http.Client) *SearchClient {
	s.http = c
	return s
}

// Search queries the Mintlify Search API and returns structured results.
func (s *SearchClient) Search(ctx context.Context, opts SearchOptions) (*SearchResult, error) {
	if s.endpoint == "" || s.apiKey == "" {
		return nil, ErrNoSearchConfig
	}

	u, err := url.Parse(s.endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing search endpoint: %w", err)
	}

	q := u.Query()
	q.Set("query", opts.Query)
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Language != "" {
		q.Set("language", opts.Language)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading search response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned %d: %s", resp.StatusCode, string(body))
	}

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}
	return &result, nil
}

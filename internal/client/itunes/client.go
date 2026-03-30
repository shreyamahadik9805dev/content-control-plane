package itunes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultLimit = 50

type Client struct {
	baseURL    string
	httpClient *http.Client
	mock       bool
}

// New builds a client; empty baseURL falls back to Apple's public host, mock skips HTTP entirely.
func New(baseURL string, mock bool) *Client {
	u := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if u == "" {
		u = "https://itunes.apple.com"
	}
	return &Client{
		baseURL: u,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		mock: mock,
	}
}

type Show struct {
	CollectionID   int64    `json:"collectionId"`
	ArtistName     string   `json:"artistName"`
	CollectionName string   `json:"collectionName"`
	FeedURL        string   `json:"feedUrl"`
	ArtworkURL600  string   `json:"artworkUrl600"`
	Genres         []string `json:"genres"`
	TrackCount     int      `json:"trackCount"`
}

type searchResponse struct {
	ResultCount int    `json:"resultCount"`
	Results     []Show `json:"results"`
}

// SearchPodcasts hits /search?media=podcast with a few retries on flaky networks (mock path is instant).
func (c *Client) SearchPodcasts(ctx context.Context, term string) ([]Show, error) {
	term = strings.TrimSpace(term)
	if c.mock {
		return mockShows(term), nil
	}

	u, err := url.Parse(c.baseURL + "/search")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("term", term)
	q.Set("media", "podcast")
	q.Set("limit", strconv.Itoa(defaultLimit))
	u.RawQuery = q.Encode()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(200*(1<<attempt)) * time.Millisecond):
			}
		}
		shows, err := c.fetchOnce(ctx, u.String())
		if err == nil {
			return shows, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("itunes search: after retries: %w", lastErr)
}

// fetchOnce does a single GET and JSON decode (8MB body cap so we don't OOM on garbage).
func (c *Client) fetchOnce(ctx context.Context, rawURL string) ([]Show, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ContentControlPlane/0.1 (github.com/shreyafeo/content-control-plane)")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		snippet := strings.TrimSpace(string(body[:min(200, len(body))]))
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, snippet)
	}

	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed.Results, nil
}

// min caps how much of a bad response body we stuff into an error string.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// mockShows fabricates two stable-looking podcasts so CI and airplane mode still work.
func mockShows(term string) []Show {
	t := strings.TrimSpace(term)
	if t == "" {
		t = "demo"
	}
	tc1 := 142
	tc2 := 89
	return []Show{
		{
			CollectionID:   1000000001,
			ArtistName:     "Demo Network",
			CollectionName: "The " + t + " hour (mock)",
			FeedURL:        "https://example.com/feeds/mock-" + url.PathEscape(t) + ".xml",
			ArtworkURL600:  "https://placehold.co/600x600/1a1f2e/3ee0c7?text=Pod",
			Genres:         []string{"Technology", "Podcasts"},
			TrackCount:     tc1,
		},
		{
			CollectionID:   1000000002,
			ArtistName:     "Indie " + t,
			CollectionName: "Deep Cuts: " + t,
			FeedURL:        "https://example.com/feeds/deep-" + url.PathEscape(t) + ".xml",
			ArtworkURL600:  "https://placehold.co/600x600/2a3142/f0b429?text=RSS",
			Genres:         []string{"Society & Culture", "Podcasts"},
			TrackCount:     tc2,
		},
	}
}

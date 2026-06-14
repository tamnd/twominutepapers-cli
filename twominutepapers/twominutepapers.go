// Package twominutepapers is the library behind the twominutepapers command: the
// HTTP client, request shaping, and the typed data models for the Two Minute
// Papers YouTube channel.
//
// The client fetches the public YouTube Atom RSS feed at
// https://www.youtube.com/feeds/videos.xml?channel_id=UCbfYPyITQ-7l4upoX8nvctg
// No authentication is required. It sets a real User-Agent, paces requests,
// and retries transient 429/5xx errors with exponential back-off.
package twominutepapers

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Config holds constructor parameters.
type Config struct {
	FeedURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible production defaults.
func DefaultConfig() Config {
	return Config{
		FeedURL:   "https://www.youtube.com/feeds/videos.xml?channel_id=UCbfYPyITQ-7l4upoX8nvctg",
		UserAgent: "Mozilla/5.0 (compatible; twominutepapers-cli/dev; +https://github.com/tamnd/twominutepapers-cli)",
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to the Two Minute Papers YouTube Atom feed.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// atomFeed is the top-level Atom envelope.
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []atomEntry `xml:"entry"`
}

// atomEntry is one video entry from the Atom feed.
// Go's encoding/xml matches namespace-prefixed elements by local name,
// so xml:"videoId" matches <yt:videoId> and xml:"group" matches <media:group>.
type atomEntry struct {
	Title   string `xml:"title"`
	VideoID string `xml:"videoId"`
	Link    struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Published string `xml:"published"`
	Group     struct {
		Thumbnail struct {
			URL string `xml:"url,attr"`
		} `xml:"thumbnail"`
		Description string `xml:"description"`
	} `xml:"group"`
}

// Videos fetches the latest videos from the Two Minute Papers Atom feed.
// It returns at most limit items (all items if limit <= 0).
func (c *Client) Videos(ctx context.Context, limit int) ([]Video, error) {
	raw, err := c.get(ctx, c.cfg.FeedURL)
	if err != nil {
		return nil, fmt.Errorf("videos: %w", err)
	}

	var feed atomFeed
	if err := xml.Unmarshal(raw, &feed); err != nil {
		return nil, fmt.Errorf("videos: parse feed: %w", err)
	}

	entries := feed.Entries
	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	out := make([]Video, 0, len(entries))
	for i, e := range entries {
		out = append(out, Video{
			Rank:        i + 1,
			Title:       strings.TrimSpace(e.Title),
			VideoID:     strings.TrimSpace(e.VideoID),
			URL:         strings.TrimSpace(e.Link.Href),
			PublishedAt: parsePublished(e.Published),
			Thumbnail:   strings.TrimSpace(e.Group.Thumbnail.URL),
			Description: truncateDesc(strings.TrimSpace(e.Group.Description), 200),
		})
	}
	return out, nil
}

// get issues a GET request with retry logic.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, url)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}

// parsePublished parses an RFC3339 timestamp and returns "YYYY-MM-DD".
// Falls back to the raw string if parsing fails.
func parsePublished(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return strings.TrimSpace(s)
	}
	return t.Format("2006-01-02")
}

// truncateDesc truncates s to at most n runes.
func truncateDesc(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n])
}

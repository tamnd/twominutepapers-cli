package twominutepapers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/twominutepapers-cli/twominutepapers"
)

const mockFeed = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns:yt="http://www.youtube.com/xml/schemas/2015" xmlns:media="http://search.yahoo.com/mrss/" xmlns="http://www.w3.org/2005/Atom">
  <title>Two Minute Papers</title>
  <entry>
    <id>yt:video:abc123</id>
    <yt:videoId>abc123</yt:videoId>
    <title>Test Video One</title>
    <link rel="alternate" href="https://www.youtube.com/watch?v=abc123"/>
    <published>2026-06-01T10:00:00+00:00</published>
    <media:group>
      <media:thumbnail url="https://i1.ytimg.com/vi/abc123/hqdefault.jpg" width="480" height="360"/>
      <media:description>First test video description.</media:description>
    </media:group>
  </entry>
  <entry>
    <id>yt:video:def456</id>
    <yt:videoId>def456</yt:videoId>
    <title>Test Video Two</title>
    <link rel="alternate" href="https://www.youtube.com/watch?v=def456"/>
    <published>2026-05-31T10:00:00+00:00</published>
    <media:group>
      <media:thumbnail url="https://i1.ytimg.com/vi/def456/hqdefault.jpg" width="480" height="360"/>
      <media:description>Second test video description.</media:description>
    </media:group>
  </entry>
</feed>`

func newTestClient(ts *httptest.Server) *twominutepapers.Client {
	cfg := twominutepapers.DefaultConfig()
	cfg.FeedURL = ts.URL + "/feed"
	cfg.Rate = 0
	return twominutepapers.NewClient(cfg)
}

func TestVideosSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = fmt.Fprint(w, mockFeed)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Videos(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVideosParsesItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, mockFeed)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	videos, err := c.Videos(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 2 {
		t.Fatalf("got %d videos, want 2", len(videos))
	}

	v := videos[0]
	if v.Rank != 1 {
		t.Errorf("rank = %d, want 1", v.Rank)
	}
	if v.Title != "Test Video One" {
		t.Errorf("title = %q", v.Title)
	}
	if v.VideoID != "abc123" {
		t.Errorf("video_id = %q, want abc123", v.VideoID)
	}
	if v.URL != "https://www.youtube.com/watch?v=abc123" {
		t.Errorf("url = %q", v.URL)
	}
	if v.PublishedAt != "2026-06-01" {
		t.Errorf("published_at = %q, want 2026-06-01", v.PublishedAt)
	}
}

func TestVideosLimitRespected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, mockFeed)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	videos, err := c.Videos(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 1 {
		t.Fatalf("got %d videos, want 1", len(videos))
	}
}

func TestVideosRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, mockFeed)
	}))
	defer srv.Close()

	cfg := twominutepapers.DefaultConfig()
	cfg.FeedURL = srv.URL + "/feed"
	cfg.Rate = 0
	cfg.Retries = 5
	c := twominutepapers.NewClient(cfg)

	start := time.Now()
	_, err := c.Videos(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

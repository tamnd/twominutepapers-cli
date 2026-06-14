package twominutepapers

// Video is one entry from the Two Minute Papers YouTube channel Atom feed.
type Video struct {
	Rank        int    `json:"rank"`
	Title       string `json:"title"`
	VideoID     string `json:"video_id"`
	URL         string `json:"url"`
	PublishedAt string `json:"published_at"` // YYYY-MM-DD
	Thumbnail   string `json:"thumbnail"`
	Description string `json:"description"` // truncated to 200 chars
}

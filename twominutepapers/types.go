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

// ChannelInfo holds aggregate statistics for the channel feed.
type ChannelInfo struct {
	TotalVideos int    `json:"total_videos"`
	OldestVideo string `json:"oldest_video"`
	LatestVideo string `json:"latest_video"`
	FeedURL     string `json:"feed_url"`
	ChannelURL  string `json:"channel_url"`
}

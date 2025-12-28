package model

import "time"

// Video represents a YouTube video with its metadata.
type Video struct {
	ID          string
	Title       string
	Description string
	ViewCount   int64
	LikeCount   int64
	Channel     string
	ChannelID   string
	PublishedAt time.Time
	Duration    int // seconds
}

// IsShort returns true if the video is 60 seconds or less (YouTube Shorts format).
func (v *Video) IsShort() bool {
	return v.Duration <= 60
}

// EngagementRate calculates the like-to-view ratio as a percentage.
func (v *Video) EngagementRate() float64 {
	if v.ViewCount == 0 {
		return 0.0
	}
	return float64(v.LikeCount) / float64(v.ViewCount) * 100
}

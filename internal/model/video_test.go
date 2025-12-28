package model

import (
	"testing"
	"time"
)

func TestVideo_Fields(t *testing.T) {
	publishedAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	v := Video{
		ID:          "abc123",
		Title:       "Test Video Title",
		Description: "A test video description",
		ViewCount:   1000000,
		LikeCount:   50000,
		Channel:     "TestChannel",
		ChannelID:   "UC123",
		PublishedAt: publishedAt,
		Duration:    60,
	}

	if v.ID != "abc123" {
		t.Errorf("ID = %q, want %q", v.ID, "abc123")
	}
	if v.Title != "Test Video Title" {
		t.Errorf("Title = %q, want %q", v.Title, "Test Video Title")
	}
	if v.Description != "A test video description" {
		t.Errorf("Description = %q, want %q", v.Description, "A test video description")
	}
	if v.ViewCount != 1000000 {
		t.Errorf("ViewCount = %d, want %d", v.ViewCount, 1000000)
	}
	if v.LikeCount != 50000 {
		t.Errorf("LikeCount = %d, want %d", v.LikeCount, 50000)
	}
	if v.Channel != "TestChannel" {
		t.Errorf("Channel = %q, want %q", v.Channel, "TestChannel")
	}
	if v.ChannelID != "UC123" {
		t.Errorf("ChannelID = %q, want %q", v.ChannelID, "UC123")
	}
	if !v.PublishedAt.Equal(publishedAt) {
		t.Errorf("PublishedAt = %v, want %v", v.PublishedAt, publishedAt)
	}
	if v.Duration != 60 {
		t.Errorf("Duration = %d, want %d", v.Duration, 60)
	}
}

func TestVideo_IsShort(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		want     bool
	}{
		{"under 60 seconds", 45, true},
		{"exactly 60 seconds", 60, true},
		{"over 60 seconds", 61, false},
		{"zero duration", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Video{Duration: tt.duration}
			if got := v.IsShort(); got != tt.want {
				t.Errorf("IsShort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVideo_EngagementRate(t *testing.T) {
	tests := []struct {
		name      string
		viewCount int64
		likeCount int64
		want      float64
	}{
		{"normal engagement", 1000, 50, 5.0},
		{"high engagement", 100, 10, 10.0},
		{"zero views", 0, 10, 0.0},
		{"zero likes", 1000, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Video{ViewCount: tt.viewCount, LikeCount: tt.likeCount}
			got := v.EngagementRate()
			if got != tt.want {
				t.Errorf("EngagementRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

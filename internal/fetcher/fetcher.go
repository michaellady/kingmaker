// Package fetcher orchestrates the Shorts fetching pipeline:
// search → get details → verify shorts → return verified videos.
package fetcher

import (
	"context"
	"errors"

	"github.com/mikelady/kingmaker/internal/model"
)

// YouTubeClient defines the interface for YouTube API operations.
type YouTubeClient interface {
	Search(ctx context.Context, query string, maxResults int64) ([]model.Video, error)
	GetVideoDetails(ctx context.Context, videoIDs []string) ([]model.Video, error)
	QuotaUsed() int64
}

// ShortsChecker defines the interface for verifying YouTube Shorts.
type ShortsChecker interface {
	IsShort(ctx context.Context, videoID string) (bool, error)
	CheckBatch(ctx context.Context, videoIDs []string) (map[string]bool, error)
}

// ShortsFetcher defines the interface for fetching verified Shorts.
type ShortsFetcher interface {
	FetchShorts(ctx context.Context, query string, maxResults int64) ([]model.Video, error)
}

// Fetcher orchestrates the Shorts fetching pipeline.
type Fetcher struct {
	youtube YouTubeClient
	shorts  ShortsChecker
}

// New creates a new Fetcher with the given YouTube client and Shorts checker.
func New(youtube YouTubeClient, shorts ShortsChecker) *Fetcher {
	return &Fetcher{
		youtube: youtube,
		shorts:  shorts,
	}
}

// FetchShorts searches for videos, gets their details, verifies they are Shorts,
// and returns only the verified Shorts.
//
// Pipeline:
// 1. Search YouTube with query (videoDuration=short filter)
// 2. Extract video IDs from search results
// 3. Verify each video is a Short via URL redirect check
// 4. Return only verified Shorts with full metadata
func (f *Fetcher) FetchShorts(ctx context.Context, query string, maxResults int64) ([]model.Video, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	if maxResults <= 0 {
		return nil, errors.New("maxResults must be positive")
	}

	// Step 1: Search for short videos
	// Note: YouTube's videoDuration=short returns videos <4 min, not just Shorts
	videos, err := f.youtube.Search(ctx, query, maxResults)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		return []model.Video{}, nil
	}

	// Step 2: Extract video IDs
	videoIDs := make([]string, len(videos))
	videoMap := make(map[string]model.Video)
	for i, v := range videos {
		videoIDs[i] = v.ID
		videoMap[v.ID] = v
	}

	// Step 3: Verify which videos are actual Shorts
	shortsStatus, err := f.shorts.CheckBatch(ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	// Step 4: Filter to only verified Shorts
	var verifiedShorts []model.Video
	for _, id := range videoIDs {
		if shortsStatus[id] {
			verifiedShorts = append(verifiedShorts, videoMap[id])
		}
	}

	return verifiedShorts, nil
}

// Package youtube provides a client for the YouTube Data API v3.
// It wraps search.list and videos.list with quota tracking.
package youtube

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/mikelady/kingmaker/internal/model"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Quota costs per API call (as of YouTube Data API v3)
const (
	QuotaCostSearch = 100 // search.list costs 100 units
	QuotaCostVideos = 1   // videos.list costs 1 unit
	MaxVideosPerRequest = 50 // Maximum video IDs per videos.list call
)

// YouTubeClient defines the interface for YouTube API operations.
type YouTubeClient interface {
	// Search finds videos matching the query with videoDuration=short filter.
	Search(ctx context.Context, query string, maxResults int64) ([]model.Video, error)

	// GetVideoDetails fetches detailed information for the given video IDs.
	// Automatically batches requests for more than 50 IDs.
	GetVideoDetails(ctx context.Context, videoIDs []string) ([]model.Video, error)

	// QuotaUsed returns the total quota units consumed by this client.
	QuotaUsed() int64
}

// YouTubeService abstracts the YouTube API for testing.
type YouTubeService interface {
	SearchList(ctx context.Context, query string, maxResults int64) (*youtube.SearchListResponse, error)
	VideosList(ctx context.Context, ids []string) (*youtube.VideoListResponse, error)
}

// Client implements YouTubeClient using the official YouTube API.
type Client struct {
	service   YouTubeService
	quotaUsed int64
}

// realYouTubeService wraps the actual YouTube API service.
type realYouTubeService struct {
	svc *youtube.Service
}

func (r *realYouTubeService) SearchList(ctx context.Context, query string, maxResults int64) (*youtube.SearchListResponse, error) {
	call := r.svc.Search.List([]string{"id"}).
		Context(ctx).
		Q(query).
		Type("video").
		VideoDuration("short"). // Filter for short videos (<4 min)
		MaxResults(maxResults).
		Order("viewCount")

	return call.Do()
}

func (r *realYouTubeService) VideosList(ctx context.Context, ids []string) (*youtube.VideoListResponse, error) {
	call := r.svc.Videos.List([]string{"snippet", "statistics", "contentDetails"}).
		Context(ctx).
		Id(ids...)

	return call.Do()
}

// NewClient creates a new YouTube API client with the given API key.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}

	ctx := context.Background()
	svc, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &Client{
		service: &realYouTubeService{svc: svc},
	}, nil
}

// Search finds videos matching the query using search.list with videoDuration=short.
// Returns basic video info; use GetVideoDetails for full metadata.
func (c *Client) Search(ctx context.Context, query string, maxResults int64) ([]model.Video, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	if maxResults <= 0 {
		return nil, errors.New("maxResults must be positive")
	}

	// Execute search
	resp, err := c.service.SearchList(ctx, query, maxResults)
	if err != nil {
		return nil, err
	}

	// Track quota
	atomic.AddInt64(&c.quotaUsed, QuotaCostSearch)

	// Extract video IDs
	videoIDs := make([]string, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item.Id != nil && item.Id.VideoId != "" {
			videoIDs = append(videoIDs, item.Id.VideoId)
		}
	}

	if len(videoIDs) == 0 {
		return []model.Video{}, nil
	}

	// Fetch full video details
	return c.GetVideoDetails(ctx, videoIDs)
}

// GetVideoDetails fetches detailed information for the given video IDs.
// Automatically batches requests for more than 50 IDs.
func (c *Client) GetVideoDetails(ctx context.Context, videoIDs []string) ([]model.Video, error) {
	if len(videoIDs) == 0 {
		return []model.Video{}, nil
	}

	var allVideos []model.Video

	// Process in batches of 50
	for i := 0; i < len(videoIDs); i += MaxVideosPerRequest {
		end := i + MaxVideosPerRequest
		if end > len(videoIDs) {
			end = len(videoIDs)
		}
		batch := videoIDs[i:end]

		resp, err := c.service.VideosList(ctx, batch)
		if err != nil {
			return nil, err
		}

		// Track quota
		atomic.AddInt64(&c.quotaUsed, QuotaCostVideos)

		// Convert to model.Video
		for _, item := range resp.Items {
			allVideos = append(allVideos, convertVideo(item))
		}
	}

	return allVideos, nil
}

// QuotaUsed returns the total quota units consumed by this client.
func (c *Client) QuotaUsed() int64 {
	return atomic.LoadInt64(&c.quotaUsed)
}

// convertVideo converts a YouTube API Video to our model.Video.
func convertVideo(v *youtube.Video) model.Video {
	video := model.Video{
		ID: v.Id,
	}

	if v.Snippet != nil {
		video.Title = v.Snippet.Title
		video.Description = v.Snippet.Description
		video.Channel = v.Snippet.ChannelTitle
		video.ChannelID = v.Snippet.ChannelId

		if v.Snippet.PublishedAt != "" {
			if t, err := time.Parse(time.RFC3339, v.Snippet.PublishedAt); err == nil {
				video.PublishedAt = t
			}
		}
	}

	if v.Statistics != nil {
		video.ViewCount = int64(v.Statistics.ViewCount)
		video.LikeCount = int64(v.Statistics.LikeCount)
	}

	if v.ContentDetails != nil {
		video.Duration = parseDuration(v.ContentDetails.Duration)
	}

	return video
}

// parseDuration converts ISO 8601 duration (e.g., "PT1M30S") to seconds.
var durationRegex = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

func parseDuration(iso8601 string) int {
	if iso8601 == "" {
		return 0
	}

	matches := durationRegex.FindStringSubmatch(iso8601)
	if matches == nil {
		return 0
	}

	var seconds int

	if matches[1] != "" {
		h, _ := strconv.Atoi(matches[1])
		seconds += h * 3600
	}
	if matches[2] != "" {
		m, _ := strconv.Atoi(matches[2])
		seconds += m * 60
	}
	if matches[3] != "" {
		s, _ := strconv.Atoi(matches[3])
		seconds += s
	}

	return seconds
}

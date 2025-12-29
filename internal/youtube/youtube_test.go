package youtube

import (
	"context"
	"testing"
	"time"

	"github.com/mikelady/kingmaker/internal/model"
	"google.golang.org/api/youtube/v3"
)

// mockYouTubeService implements YouTubeService for testing
type mockYouTubeService struct {
	searchResults  *youtube.SearchListResponse
	searchErr      error
	videosResults  *youtube.VideoListResponse
	videosErr      error
	searchCalls    int
	videosCalls    int
}

func (m *mockYouTubeService) SearchList(ctx context.Context, query string, maxResults int64) (*youtube.SearchListResponse, error) {
	return m.SearchListWithDuration(ctx, query, maxResults, DurationShort)
}

func (m *mockYouTubeService) SearchListWithDuration(ctx context.Context, query string, maxResults int64, duration string) (*youtube.SearchListResponse, error) {
	m.searchCalls++
	return m.searchResults, m.searchErr
}

func (m *mockYouTubeService) VideosList(ctx context.Context, ids []string) (*youtube.VideoListResponse, error) {
	m.videosCalls++
	return m.videosResults, m.videosErr
}

func TestNewClient(t *testing.T) {
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
	if client.QuotaUsed() != 0 {
		t.Errorf("Initial quota should be 0, got %d", client.QuotaUsed())
	}
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Error("expected error for empty API key")
	}
}

func TestSearch_ReturnsVideos(t *testing.T) {
	mock := &mockYouTubeService{
		searchResults: &youtube.SearchListResponse{
			Items: []*youtube.SearchResult{
				{Id: &youtube.ResourceId{VideoId: "vid1"}},
				{Id: &youtube.ResourceId{VideoId: "vid2"}},
			},
		},
		videosResults: &youtube.VideoListResponse{
			Items: []*youtube.Video{
				{
					Id: "vid1",
					Snippet: &youtube.VideoSnippet{
						Title:       "Test Video 1",
						Description: "Description 1",
						ChannelId:   "chan1",
						ChannelTitle: "Channel 1",
						PublishedAt: "2024-01-15T10:00:00Z",
					},
					Statistics: &youtube.VideoStatistics{
						ViewCount: 1000,
						LikeCount: 100,
					},
					ContentDetails: &youtube.VideoContentDetails{
						Duration: "PT45S", // 45 seconds
					},
				},
				{
					Id: "vid2",
					Snippet: &youtube.VideoSnippet{
						Title:       "Test Video 2",
						Description: "Description 2",
						ChannelId:   "chan2",
						ChannelTitle: "Channel 2",
						PublishedAt: "2024-01-16T10:00:00Z",
					},
					Statistics: &youtube.VideoStatistics{
						ViewCount: 2000,
						LikeCount: 200,
					},
					ContentDetails: &youtube.VideoContentDetails{
						Duration: "PT30S", // 30 seconds
					},
				},
			},
		},
	}

	client := &Client{service: mock}
	videos, err := client.Search(context.Background(), "test query", 10)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(videos) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(videos))
	}

	// Check first video
	if videos[0].ID != "vid1" {
		t.Errorf("expected ID 'vid1', got '%s'", videos[0].ID)
	}
	if videos[0].Title != "Test Video 1" {
		t.Errorf("expected title 'Test Video 1', got '%s'", videos[0].Title)
	}
	if videos[0].ViewCount != 1000 {
		t.Errorf("expected viewCount 1000, got %d", videos[0].ViewCount)
	}
	if videos[0].Duration != 45 {
		t.Errorf("expected duration 45s, got %d", videos[0].Duration)
	}

	// Check quota tracking
	// search = 100, videos = 1
	if client.QuotaUsed() != 101 {
		t.Errorf("expected quota 101, got %d", client.QuotaUsed())
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	client := &Client{service: &mockYouTubeService{}}
	_, err := client.Search(context.Background(), "", 10)

	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestSearch_InvalidMaxResults(t *testing.T) {
	client := &Client{service: &mockYouTubeService{}}
	_, err := client.Search(context.Background(), "test", 0)

	if err == nil {
		t.Error("expected error for invalid maxResults")
	}

	_, err = client.Search(context.Background(), "test", -1)
	if err == nil {
		t.Error("expected error for negative maxResults")
	}
}

func TestGetVideoDetails_SingleVideo(t *testing.T) {
	mock := &mockYouTubeService{
		videosResults: &youtube.VideoListResponse{
			Items: []*youtube.Video{
				{
					Id: "test123",
					Snippet: &youtube.VideoSnippet{
						Title:       "Test Title",
						Description: "Test Description",
						ChannelId:   "chan1",
						ChannelTitle: "Test Channel",
						PublishedAt: "2024-06-15T12:30:00Z",
					},
					Statistics: &youtube.VideoStatistics{
						ViewCount: 5000,
						LikeCount: 500,
					},
					ContentDetails: &youtube.VideoContentDetails{
						Duration: "PT1M", // 60 seconds
					},
				},
			},
		},
	}

	client := &Client{service: mock}
	videos, err := client.GetVideoDetails(context.Background(), []string{"test123"})

	if err != nil {
		t.Fatalf("GetVideoDetails failed: %v", err)
	}
	if len(videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(videos))
	}
	if videos[0].ID != "test123" {
		t.Errorf("expected ID 'test123', got '%s'", videos[0].ID)
	}
	if videos[0].Duration != 60 {
		t.Errorf("expected duration 60s, got %d", videos[0].Duration)
	}
}

func TestGetVideoDetails_EmptyInput(t *testing.T) {
	client := &Client{service: &mockYouTubeService{}}
	videos, err := client.GetVideoDetails(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(videos) != 0 {
		t.Errorf("expected 0 videos, got %d", len(videos))
	}
}

func TestGetVideoDetails_BatchesOver50(t *testing.T) {
	// Create 60 video IDs
	ids := make([]string, 60)
	for i := 0; i < 60; i++ {
		ids[i] = "vid" + string(rune('a'+i%26))
	}

	callCount := 0
	mock := &mockYouTubeService{
		videosResults: &youtube.VideoListResponse{
			Items: []*youtube.Video{},
		},
	}

	// Override to track calls
	client := &Client{service: mock}
	_, _ = client.GetVideoDetails(context.Background(), ids)

	// Should have made 2 calls (50 + 10)
	if mock.videosCalls != 2 {
		t.Errorf("expected 2 API calls for 60 videos, got %d", callCount)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"PT45S", 45},
		{"PT1M", 60},
		{"PT1M30S", 90},
		{"PT2M", 120},
		{"PT1H", 3600},
		{"PT1H30M45S", 5445},
		{"PT0S", 0},
		{"", 0},
		{"invalid", 0},
	}

	for _, tc := range tests {
		result := parseDuration(tc.input)
		if result != tc.expected {
			t.Errorf("parseDuration(%q) = %d, want %d", tc.input, result, tc.expected)
		}
	}
}

func TestConvertVideo(t *testing.T) {
	ytVideo := &youtube.Video{
		Id: "xyz789",
		Snippet: &youtube.VideoSnippet{
			Title:        "Amazing Video",
			Description:  "Great content #trending",
			ChannelId:    "UCtest",
			ChannelTitle: "Test Channel",
			PublishedAt:  "2024-03-20T15:00:00Z",
		},
		Statistics: &youtube.VideoStatistics{
			ViewCount: 10000,
			LikeCount: 1000,
		},
		ContentDetails: &youtube.VideoContentDetails{
			Duration: "PT55S",
		},
	}

	video := convertVideo(ytVideo)

	if video.ID != "xyz789" {
		t.Errorf("ID = %s, want xyz789", video.ID)
	}
	if video.Title != "Amazing Video" {
		t.Errorf("Title = %s, want 'Amazing Video'", video.Title)
	}
	if video.Channel != "Test Channel" {
		t.Errorf("Channel = %s, want 'Test Channel'", video.Channel)
	}
	if video.ChannelID != "UCtest" {
		t.Errorf("ChannelID = %s, want 'UCtest'", video.ChannelID)
	}
	if video.ViewCount != 10000 {
		t.Errorf("ViewCount = %d, want 10000", video.ViewCount)
	}
	if video.LikeCount != 1000 {
		t.Errorf("LikeCount = %d, want 1000", video.LikeCount)
	}
	if video.Duration != 55 {
		t.Errorf("Duration = %d, want 55", video.Duration)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2024-03-20T15:00:00Z")
	if !video.PublishedAt.Equal(expectedTime) {
		t.Errorf("PublishedAt = %v, want %v", video.PublishedAt, expectedTime)
	}
}

func TestClient_Interface(t *testing.T) {
	// Verify Client implements YouTubeClient interface
	var _ YouTubeClient = (*Client)(nil)
}

func TestQuotaCosts(t *testing.T) {
	if QuotaCostSearch != 100 {
		t.Errorf("QuotaCostSearch = %d, want 100", QuotaCostSearch)
	}
	if QuotaCostVideos != 1 {
		t.Errorf("QuotaCostVideos = %d, want 1", QuotaCostVideos)
	}
}

func TestVideoResult(t *testing.T) {
	// Verify model.Video is returned correctly
	v := model.Video{
		ID:          "test",
		Title:       "Test",
		Description: "Desc",
		ViewCount:   100,
		LikeCount:   10,
		Channel:     "Chan",
		ChannelID:   "ChanID",
		Duration:    45,
	}

	if !v.IsShort() {
		t.Error("45s video should be a Short")
	}
}

// Tests for SearchWithDuration (91w.9)

// mockConfigurableService tracks the duration filter used
type mockConfigurableService struct {
	searchResults      *youtube.SearchListResponse
	searchErr          error
	videosResults      *youtube.VideoListResponse
	videosErr          error
	searchCalls        int
	videosCalls        int
	lastDurationFilter string
}

func (m *mockConfigurableService) SearchList(ctx context.Context, query string, maxResults int64) (*youtube.SearchListResponse, error) {
	return m.SearchListWithDuration(ctx, query, maxResults, DurationShort)
}

func (m *mockConfigurableService) SearchListWithDuration(ctx context.Context, query string, maxResults int64, duration string) (*youtube.SearchListResponse, error) {
	m.lastDurationFilter = duration
	m.searchCalls++
	return m.searchResults, m.searchErr
}

func (m *mockConfigurableService) VideosList(ctx context.Context, ids []string) (*youtube.VideoListResponse, error) {
	m.videosCalls++
	return m.videosResults, m.videosErr
}

func TestSearchWithDuration_Short(t *testing.T) {
	mock := &mockConfigurableService{
		searchResults: &youtube.SearchListResponse{
			Items: []*youtube.SearchResult{
				{Id: &youtube.ResourceId{VideoId: "vid1"}},
			},
		},
		videosResults: &youtube.VideoListResponse{
			Items: []*youtube.Video{
				{
					Id:      "vid1",
					Snippet: &youtube.VideoSnippet{Title: "Test"},
				},
			},
		},
	}

	client := &Client{service: mock}
	_, err := client.SearchWithDuration(context.Background(), "test", 10, DurationShort)

	if err != nil {
		t.Fatalf("SearchWithDuration failed: %v", err)
	}
	if mock.lastDurationFilter != "short" {
		t.Errorf("expected duration filter 'short', got '%s'", mock.lastDurationFilter)
	}
}

func TestSearchWithDuration_Medium(t *testing.T) {
	mock := &mockConfigurableService{
		searchResults: &youtube.SearchListResponse{Items: []*youtube.SearchResult{}},
	}

	client := &Client{service: mock}
	_, err := client.SearchWithDuration(context.Background(), "test", 10, DurationMedium)

	if err != nil {
		t.Fatalf("SearchWithDuration failed: %v", err)
	}
	if mock.lastDurationFilter != "medium" {
		t.Errorf("expected duration filter 'medium', got '%s'", mock.lastDurationFilter)
	}
}

func TestSearchWithDuration_Long(t *testing.T) {
	mock := &mockConfigurableService{
		searchResults: &youtube.SearchListResponse{Items: []*youtube.SearchResult{}},
	}

	client := &Client{service: mock}
	_, err := client.SearchWithDuration(context.Background(), "test", 10, DurationLong)

	if err != nil {
		t.Fatalf("SearchWithDuration failed: %v", err)
	}
	if mock.lastDurationFilter != "long" {
		t.Errorf("expected duration filter 'long', got '%s'", mock.lastDurationFilter)
	}
}

func TestSearchWithDuration_Any(t *testing.T) {
	mock := &mockConfigurableService{
		searchResults: &youtube.SearchListResponse{Items: []*youtube.SearchResult{}},
	}

	client := &Client{service: mock}
	_, err := client.SearchWithDuration(context.Background(), "test", 10, DurationAny)

	if err != nil {
		t.Fatalf("SearchWithDuration failed: %v", err)
	}
	if mock.lastDurationFilter != "" {
		t.Errorf("expected no duration filter, got '%s'", mock.lastDurationFilter)
	}
}

func TestSearch_UsesShortByDefault(t *testing.T) {
	// Verify original Search() still uses "short" filter for backward compat
	mock := &mockConfigurableService{
		searchResults: &youtube.SearchListResponse{Items: []*youtube.SearchResult{}},
	}

	client := &Client{service: mock}
	_, err := client.Search(context.Background(), "test", 10)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	// Search calls SearchWithDuration with DurationShort
	if mock.lastDurationFilter != "short" {
		t.Errorf("expected duration filter 'short' from Search(), got '%s'", mock.lastDurationFilter)
	}
}

func TestDurationConstants(t *testing.T) {
	if DurationShort != "short" {
		t.Errorf("DurationShort = %s, want 'short'", DurationShort)
	}
	if DurationMedium != "medium" {
		t.Errorf("DurationMedium = %s, want 'medium'", DurationMedium)
	}
	if DurationLong != "long" {
		t.Errorf("DurationLong = %s, want 'long'", DurationLong)
	}
	if DurationAny != "" {
		t.Errorf("DurationAny = %s, want ''", DurationAny)
	}
}

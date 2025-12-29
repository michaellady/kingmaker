package fetcher

import (
	"context"
	"errors"
	"testing"

	"github.com/mikelady/kingmaker/internal/model"
)

// Mock YouTube client
type mockYouTubeClient struct {
	searchResults []model.Video
	searchErr     error
	detailResults []model.Video
	detailErr     error
	searchCalls   int
	detailCalls   int
}

func (m *mockYouTubeClient) Search(ctx context.Context, query string, maxResults int64) ([]model.Video, error) {
	m.searchCalls++
	return m.searchResults, m.searchErr
}

func (m *mockYouTubeClient) GetVideoDetails(ctx context.Context, videoIDs []string) ([]model.Video, error) {
	m.detailCalls++
	return m.detailResults, m.detailErr
}

func (m *mockYouTubeClient) QuotaUsed() int64 {
	return 0
}

// Mock Shorts checker
type mockShortsChecker struct {
	results    map[string]bool
	err        error
	checkCalls int
}

func (m *mockShortsChecker) IsShort(ctx context.Context, videoID string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.results[videoID], nil
}

func (m *mockShortsChecker) CheckBatch(ctx context.Context, videoIDs []string) (map[string]bool, error) {
	m.checkCalls++
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]bool)
	for _, id := range videoIDs {
		result[id] = m.results[id]
	}
	return result, nil
}

func TestFetchShorts_ReturnsVerifiedShorts(t *testing.T) {
	ytClient := &mockYouTubeClient{
		searchResults: []model.Video{
			{ID: "short1", Title: "Short Video 1"},
			{ID: "short2", Title: "Short Video 2"},
			{ID: "notshort", Title: "Regular Video"},
		},
	}

	shortsChecker := &mockShortsChecker{
		results: map[string]bool{
			"short1":   true,
			"short2":   true,
			"notshort": false,
		},
	}

	fetcher := New(ytClient, shortsChecker)
	videos, err := fetcher.FetchShorts(context.Background(), "test query", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only return verified shorts
	if len(videos) != 2 {
		t.Errorf("expected 2 verified shorts, got %d", len(videos))
	}

	for _, v := range videos {
		if v.ID != "short1" && v.ID != "short2" {
			t.Errorf("unexpected video ID: %s", v.ID)
		}
	}
}

func TestFetchShorts_EmptyQuery(t *testing.T) {
	fetcher := New(&mockYouTubeClient{}, &mockShortsChecker{})
	_, err := fetcher.FetchShorts(context.Background(), "", 10)

	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestFetchShorts_InvalidMaxResults(t *testing.T) {
	fetcher := New(&mockYouTubeClient{}, &mockShortsChecker{})
	_, err := fetcher.FetchShorts(context.Background(), "test", 0)

	if err == nil {
		t.Error("expected error for invalid maxResults")
	}
}

func TestFetchShorts_SearchError(t *testing.T) {
	ytClient := &mockYouTubeClient{
		searchErr: errors.New("API error"),
	}

	fetcher := New(ytClient, &mockShortsChecker{})
	_, err := fetcher.FetchShorts(context.Background(), "test", 10)

	if err == nil {
		t.Error("expected error when search fails")
	}
}

func TestFetchShorts_NoResults(t *testing.T) {
	ytClient := &mockYouTubeClient{
		searchResults: []model.Video{},
	}

	fetcher := New(ytClient, &mockShortsChecker{})
	videos, err := fetcher.FetchShorts(context.Background(), "obscure query", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(videos) != 0 {
		t.Errorf("expected 0 videos, got %d", len(videos))
	}
}

func TestFetchShorts_AllFilteredOut(t *testing.T) {
	ytClient := &mockYouTubeClient{
		searchResults: []model.Video{
			{ID: "vid1", Title: "Video 1"},
			{ID: "vid2", Title: "Video 2"},
		},
	}

	shortsChecker := &mockShortsChecker{
		results: map[string]bool{
			"vid1": false,
			"vid2": false,
		},
	}

	fetcher := New(ytClient, shortsChecker)
	videos, err := fetcher.FetchShorts(context.Background(), "test", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(videos) != 0 {
		t.Errorf("expected 0 videos after filtering, got %d", len(videos))
	}
}

func TestFetchShorts_ShortsCheckError(t *testing.T) {
	ytClient := &mockYouTubeClient{
		searchResults: []model.Video{
			{ID: "vid1", Title: "Video 1"},
		},
	}

	shortsChecker := &mockShortsChecker{
		err: errors.New("network error"),
	}

	fetcher := New(ytClient, shortsChecker)
	_, err := fetcher.FetchShorts(context.Background(), "test", 10)

	if err == nil {
		t.Error("expected error when shorts check fails")
	}
}

func TestFetchShorts_PreservesVideoMetadata(t *testing.T) {
	ytClient := &mockYouTubeClient{
		searchResults: []model.Video{
			{
				ID:          "abc123",
				Title:       "Amazing Short",
				Description: "Great content",
				ViewCount:   1000,
				LikeCount:   100,
				Channel:     "Test Channel",
				Duration:    45,
			},
		},
	}

	shortsChecker := &mockShortsChecker{
		results: map[string]bool{"abc123": true},
	}

	fetcher := New(ytClient, shortsChecker)
	videos, err := fetcher.FetchShorts(context.Background(), "test", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(videos))
	}

	v := videos[0]
	if v.ID != "abc123" {
		t.Errorf("ID = %s, want abc123", v.ID)
	}
	if v.Title != "Amazing Short" {
		t.Errorf("Title = %s, want 'Amazing Short'", v.Title)
	}
	if v.ViewCount != 1000 {
		t.Errorf("ViewCount = %d, want 1000", v.ViewCount)
	}
	if v.Channel != "Test Channel" {
		t.Errorf("Channel = %s, want 'Test Channel'", v.Channel)
	}
}

func TestNew(t *testing.T) {
	ytClient := &mockYouTubeClient{}
	shortsChecker := &mockShortsChecker{}

	fetcher := New(ytClient, shortsChecker)

	if fetcher == nil {
		t.Fatal("New returned nil")
	}
}

func TestFetcher_Interface(t *testing.T) {
	// Verify Fetcher implements ShortsFetcher interface
	var _ ShortsFetcher = (*Fetcher)(nil)
}

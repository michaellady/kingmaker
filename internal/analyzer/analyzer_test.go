package analyzer

import (
	"testing"

	"github.com/mikelady/kingmaker/internal/hooks"
	"github.com/mikelady/kingmaker/internal/keywords"
	"github.com/mikelady/kingmaker/internal/model"
)

func TestAnalyzeVideos_Empty(t *testing.T) {
	result := AnalyzeVideos(nil)

	if result.VideoCount != 0 {
		t.Errorf("VideoCount = %d, want 0", result.VideoCount)
	}
	if len(result.TopHooks) != 0 {
		t.Errorf("TopHooks length = %d, want 0", len(result.TopHooks))
	}
	if len(result.TopKeywords) != 0 {
		t.Errorf("TopKeywords length = %d, want 0", len(result.TopKeywords))
	}
	if len(result.TopHashtags) != 0 {
		t.Errorf("TopHashtags length = %d, want 0", len(result.TopHashtags))
	}
}

func TestAnalyzeVideos_SingleVideo(t *testing.T) {
	videos := []model.Video{
		{
			ID:          "abc123",
			Title:       "How to code in 5 minutes",
			Description: "Learn coding fast #coding #tutorial #programming",
		},
	}

	result := AnalyzeVideos(videos)

	if result.VideoCount != 1 {
		t.Errorf("VideoCount = %d, want 1", result.VideoCount)
	}

	// Should extract "how" question hook
	foundHowHook := false
	for _, h := range result.TopHooks {
		if h.Type == hooks.Question && h.Pattern == "how" {
			foundHowHook = true
			break
		}
	}
	if !foundHowHook {
		t.Error("Expected to find 'how' question hook")
	}

	// Should extract hashtags
	if len(result.TopHashtags) < 1 {
		t.Error("Expected at least one hashtag")
	}
}

func TestAnalyzeVideos_MultipleVideos(t *testing.T) {
	videos := []model.Video{
		{
			ID:          "vid1",
			Title:       "How to build AI apps",
			Description: "AI tutorial #ai #coding",
		},
		{
			ID:          "vid2",
			Title:       "How I made $1000 with AI",
			Description: "Money with AI #ai #money",
		},
		{
			ID:          "vid3",
			Title:       "5 secrets to AI success",
			Description: "AI tips #ai #success",
		},
	}

	result := AnalyzeVideos(videos)

	if result.VideoCount != 3 {
		t.Errorf("VideoCount = %d, want 3", result.VideoCount)
	}

	// Should have hooks
	if len(result.TopHooks) == 0 {
		t.Error("Expected at least one hook")
	}

	// Should have keywords (ai should be prominent)
	foundAI := false
	for _, kw := range result.TopKeywords {
		if kw.Word == "ai" {
			foundAI = true
			if kw.Frequency < 3 {
				t.Errorf("Expected 'ai' frequency >= 3, got %d", kw.Frequency)
			}
			break
		}
	}
	if !foundAI {
		t.Error("Expected 'ai' keyword")
	}

	// Should have hashtags with #ai being most frequent
	if len(result.TopHashtags) == 0 {
		t.Error("Expected hashtags")
	}
	if result.TopHashtags[0].Tag != "ai" {
		t.Errorf("Expected top hashtag 'ai', got '%s'", result.TopHashtags[0].Tag)
	}
	if result.TopHashtags[0].Frequency != 3 {
		t.Errorf("Expected #ai frequency 3, got %d", result.TopHashtags[0].Frequency)
	}
}

func TestAnalyzeVideos_HashtagAggregation(t *testing.T) {
	videos := []model.Video{
		{Description: "#golang #programming"},
		{Description: "#golang #tutorial"},
		{Description: "#golang"},
	}

	result := AnalyzeVideos(videos)

	// #golang should appear 3 times and be first
	if len(result.TopHashtags) == 0 {
		t.Fatal("Expected hashtags")
	}
	if result.TopHashtags[0].Tag != "golang" {
		t.Errorf("Expected top hashtag 'golang', got '%s'", result.TopHashtags[0].Tag)
	}
	if result.TopHashtags[0].Frequency != 3 {
		t.Errorf("Expected frequency 3, got %d", result.TopHashtags[0].Frequency)
	}
}

func TestAnalyzeVideosWithOptions(t *testing.T) {
	videos := []model.Video{
		{Title: "Test video", Description: "#a #b #c #d #e #f #g #h #i #j #k"},
	}

	opts := Options{
		TopKeywordsN:  5,
		TopHashtagsN:  3,
	}

	result := AnalyzeVideosWithOptions(videos, opts)

	if len(result.TopHashtags) > 3 {
		t.Errorf("Expected max 3 hashtags, got %d", len(result.TopHashtags))
	}
	if len(result.TopKeywords) > 5 {
		t.Errorf("Expected max 5 keywords, got %d", len(result.TopKeywords))
	}
}

func TestPatterns_Type(t *testing.T) {
	// Verify Patterns struct has expected fields
	p := Patterns{
		TopHooks:     []hooks.Hook{},
		TopKeywords:  []keywords.Keyword{},
		TopHashtags:  []Hashtag{},
		VideoCount:   0,
	}

	if p.VideoCount != 0 {
		t.Error("Patterns struct field access failed")
	}
}

func TestHashtag_Type(t *testing.T) {
	h := Hashtag{Tag: "test", Frequency: 5}
	if h.Tag != "test" || h.Frequency != 5 {
		t.Error("Hashtag struct field access failed")
	}
}

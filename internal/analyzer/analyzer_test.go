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

// Tests for TitleMetrics (91w.3)

func TestTitleMetrics_Empty(t *testing.T) {
	result := AnalyzeVideos(nil)

	if result.TitleMetrics.AvgLength != 0 {
		t.Errorf("AvgLength = %d, want 0", result.TitleMetrics.AvgLength)
	}
	if result.TitleMetrics.MinLength != 0 {
		t.Errorf("MinLength = %d, want 0", result.TitleMetrics.MinLength)
	}
	if result.TitleMetrics.MaxLength != 0 {
		t.Errorf("MaxLength = %d, want 0", result.TitleMetrics.MaxLength)
	}
}

func TestTitleMetrics_SingleVideo(t *testing.T) {
	videos := []model.Video{
		{Title: "How to code in Go"}, // 17 chars, 5 words
	}

	result := AnalyzeVideos(videos)

	if result.TitleMetrics.AvgLength != 17 {
		t.Errorf("AvgLength = %d, want 17", result.TitleMetrics.AvgLength)
	}
	if result.TitleMetrics.MinLength != 17 {
		t.Errorf("MinLength = %d, want 17", result.TitleMetrics.MinLength)
	}
	if result.TitleMetrics.MaxLength != 17 {
		t.Errorf("MaxLength = %d, want 17", result.TitleMetrics.MaxLength)
	}
	if result.TitleMetrics.AvgWords != 5 {
		t.Errorf("AvgWords = %d, want 5", result.TitleMetrics.AvgWords)
	}
}

func TestTitleMetrics_MultipleVideos(t *testing.T) {
	videos := []model.Video{
		{Title: "Short"},           // 5 chars, 1 word
		{Title: "Medium title"},    // 12 chars, 2 words
		{Title: "This is longer"},  // 14 chars, 3 words
	}

	result := AnalyzeVideos(videos)

	// (5 + 12 + 14) / 3 = 10.33 -> 10
	if result.TitleMetrics.AvgLength != 10 {
		t.Errorf("AvgLength = %d, want 10", result.TitleMetrics.AvgLength)
	}
	if result.TitleMetrics.MinLength != 5 {
		t.Errorf("MinLength = %d, want 5", result.TitleMetrics.MinLength)
	}
	if result.TitleMetrics.MaxLength != 14 {
		t.Errorf("MaxLength = %d, want 14", result.TitleMetrics.MaxLength)
	}
	// (1 + 2 + 3) / 3 = 2
	if result.TitleMetrics.AvgWords != 2 {
		t.Errorf("AvgWords = %d, want 2", result.TitleMetrics.AvgWords)
	}
}

func TestTitleMetrics_HookDensity(t *testing.T) {
	videos := []model.Video{
		{Title: "How to build AI apps"},      // Has "how" hook
		{Title: "Why you need to learn Go"},  // Has "why" hook
		{Title: "Simple tutorial"},           // No hook
		{Title: "5 tips for success"},        // Has numerical hook
	}

	result := AnalyzeVideos(videos)

	// 3 out of 4 titles have hooks = 0.75
	if result.TitleMetrics.HookDensity < 0.74 || result.TitleMetrics.HookDensity > 0.76 {
		t.Errorf("HookDensity = %.2f, want ~0.75", result.TitleMetrics.HookDensity)
	}
}

func TestTitleMetrics_CommonPatterns(t *testing.T) {
	videos := []model.Video{
		{Title: "I built X in 5 minutes"},
		{Title: "I made Y in 10 minutes"},
		{Title: "I created Z in 1 hour"},
		{Title: "Random title"},
	}

	result := AnalyzeVideos(videos)

	// Should detect "I [verb] X in Y [time]" pattern
	foundPattern := false
	for _, p := range result.TitleMetrics.CommonPatterns {
		if p.Name == "I [verb] in [time]" {
			foundPattern = true
			if p.Count != 3 {
				t.Errorf("Pattern count = %d, want 3", p.Count)
			}
			break
		}
	}
	if !foundPattern {
		t.Error("Expected to find 'I [verb] in [time]' pattern")
	}
}

func TestTitleMetrics_EmptyTitles(t *testing.T) {
	videos := []model.Video{
		{Title: ""},
		{Title: "Valid title"},
		{Title: ""},
	}

	result := AnalyzeVideos(videos)

	// Should only count non-empty titles
	if result.TitleMetrics.AvgLength != 11 { // "Valid title" = 11 chars
		t.Errorf("AvgLength = %d, want 11", result.TitleMetrics.AvgLength)
	}
}

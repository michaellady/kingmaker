// Package analyzer aggregates patterns from video metadata using hooks and keywords extractors.
package analyzer

import (
	"regexp"
	"sort"
	"strings"

	"github.com/mikelady/kingmaker/internal/hooks"
	"github.com/mikelady/kingmaker/internal/keywords"
	"github.com/mikelady/kingmaker/internal/model"
	"github.com/mikelady/kingmaker/internal/text"
)

// Hashtag represents an extracted hashtag with its frequency.
type Hashtag struct {
	Tag       string
	Frequency int
}

// TitlePattern represents a detected title formula pattern.
type TitlePattern struct {
	Name  string  // Pattern name (e.g., "I [verb] in [time]")
	Count int     // Number of titles matching this pattern
	Ratio float64 // Proportion of titles matching
}

// TitleMetrics contains metrics about video titles for optimization.
type TitleMetrics struct {
	AvgLength      int            // Average title length in characters
	MinLength      int            // Minimum title length
	MaxLength      int            // Maximum title length
	AvgWords       int            // Average word count
	HookDensity    float64        // Proportion of titles with hooks (0.0-1.0)
	CommonPatterns []TitlePattern // Detected title formula patterns
}

// Patterns contains aggregated analysis results from video metadata.
type Patterns struct {
	TopHooks     []hooks.Hook
	TopKeywords  []keywords.Keyword
	TopHashtags  []Hashtag
	TitleMetrics TitleMetrics
	VideoCount   int
}

// Options configures the analysis behavior.
type Options struct {
	TopKeywordsN int // Number of top keywords to return (default 10)
	TopHashtagsN int // Number of top hashtags to return (default 10)
}

// DefaultOptions returns the default analysis options.
func DefaultOptions() Options {
	return Options{
		TopKeywordsN: 10,
		TopHashtagsN: 10,
	}
}

// AnalyzeVideos extracts patterns from video metadata using default options.
func AnalyzeVideos(videos []model.Video) Patterns {
	return AnalyzeVideosWithOptions(videos, DefaultOptions())
}

// AnalyzeVideosWithOptions extracts patterns from video metadata with custom options.
func AnalyzeVideosWithOptions(videos []model.Video, opts Options) Patterns {
	if len(videos) == 0 {
		return Patterns{}
	}

	// Set defaults if not specified
	if opts.TopKeywordsN <= 0 {
		opts.TopKeywordsN = 10
	}
	if opts.TopHashtagsN <= 0 {
		opts.TopHashtagsN = 10
	}

	// Extract titles and descriptions
	titles := make([]string, 0, len(videos))
	allTexts := make([]string, 0, len(videos)*2)
	descriptions := make([]string, 0, len(videos))

	for _, v := range videos {
		if v.Title != "" {
			titles = append(titles, v.Title)
			allTexts = append(allTexts, v.Title)
		}
		if v.Description != "" {
			descriptions = append(descriptions, v.Description)
			allTexts = append(allTexts, v.Description)
		}
	}

	// Extract hooks from titles
	topHooks := hooks.ExtractHooks(titles)

	// Extract keywords from all text
	topKeywords := keywords.ExtractKeywords(allTexts, opts.TopKeywordsN)

	// Extract and aggregate hashtags from descriptions
	topHashtags := extractAndAggregateHashtags(descriptions, opts.TopHashtagsN)

	// Calculate title metrics
	titleMetrics := calculateTitleMetrics(titles, topHooks)

	return Patterns{
		TopHooks:     topHooks,
		TopKeywords:  topKeywords,
		TopHashtags:  topHashtags,
		TitleMetrics: titleMetrics,
		VideoCount:   len(videos),
	}
}

// extractAndAggregateHashtags extracts hashtags from descriptions and returns top N by frequency.
func extractAndAggregateHashtags(descriptions []string, topN int) []Hashtag {
	counts := make(map[string]int)

	for _, desc := range descriptions {
		tags := text.ExtractHashtags(desc)
		for _, tag := range tags {
			counts[tag]++
		}
	}

	// Convert to slice
	hashtags := make([]Hashtag, 0, len(counts))
	for tag, freq := range counts {
		hashtags = append(hashtags, Hashtag{Tag: tag, Frequency: freq})
	}

	// Sort by frequency descending, then alphabetically
	sort.Slice(hashtags, func(i, j int) bool {
		if hashtags[i].Frequency != hashtags[j].Frequency {
			return hashtags[i].Frequency > hashtags[j].Frequency
		}
		return hashtags[i].Tag < hashtags[j].Tag
	})

	// Return top N
	if len(hashtags) > topN {
		hashtags = hashtags[:topN]
	}

	return hashtags
}

// Title pattern regexes
var (
	// "I [verb] X in Y [time]" pattern - e.g., "I built X in 5 minutes"
	iVerbInTimePattern = regexp.MustCompile(`(?i)^I\s+\w+.*\s+in\s+\d+\s*\w*$`)
)

// calculateTitleMetrics computes metrics about video titles.
func calculateTitleMetrics(titles []string, extractedHooks []hooks.Hook) TitleMetrics {
	if len(titles) == 0 {
		return TitleMetrics{}
	}

	var totalLength, totalWords int
	minLength := -1
	maxLength := 0

	for _, title := range titles {
		length := len(title)
		words := len(strings.Fields(title))

		totalLength += length
		totalWords += words

		if minLength < 0 || length < minLength {
			minLength = length
		}
		if length > maxLength {
			maxLength = length
		}
	}

	if minLength < 0 {
		minLength = 0
	}

	// Calculate hook density - proportion of titles with at least one hook
	titlesWithHooks := countTitlesWithHooks(titles)
	hookDensity := float64(titlesWithHooks) / float64(len(titles))

	// Detect common patterns
	patterns := detectTitlePatterns(titles)

	return TitleMetrics{
		AvgLength:      totalLength / len(titles),
		MinLength:      minLength,
		MaxLength:      maxLength,
		AvgWords:       totalWords / len(titles),
		HookDensity:    hookDensity,
		CommonPatterns: patterns,
	}
}

// countTitlesWithHooks counts how many titles contain at least one hook.
func countTitlesWithHooks(titles []string) int {
	count := 0
	for _, title := range titles {
		if hasHook(title) {
			count++
		}
	}
	return count
}

// hasHook checks if a title contains any hook pattern.
func hasHook(title string) bool {
	lower := strings.ToLower(title)

	// Question hooks
	questionStarters := []string{"how", "why", "what", "when", "where", "who"}
	for _, q := range questionStarters {
		if strings.HasPrefix(lower, q+" ") || strings.HasPrefix(lower, q+"\t") {
			return true
		}
	}

	// Numerical hooks (e.g., "5 tips", "10 ways")
	numericalPattern := regexp.MustCompile(`^\d+\s+\w+`)
	return numericalPattern.MatchString(lower)
}

// detectTitlePatterns identifies common title formula patterns.
func detectTitlePatterns(titles []string) []TitlePattern {
	patterns := make(map[string]int)

	for _, title := range titles {
		if iVerbInTimePattern.MatchString(title) {
			patterns["I [verb] in [time]"]++
		}
	}

	// Convert to slice and sort by count
	result := make([]TitlePattern, 0, len(patterns))
	for name, count := range patterns {
		result = append(result, TitlePattern{
			Name:  name,
			Count: count,
			Ratio: float64(count) / float64(len(titles)),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

// Package analyzer aggregates patterns from video metadata using hooks and keywords extractors.
package analyzer

import (
	"sort"

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

// Patterns contains aggregated analysis results from video metadata.
type Patterns struct {
	TopHooks    []hooks.Hook
	TopKeywords []keywords.Keyword
	TopHashtags []Hashtag
	VideoCount  int
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

	return Patterns{
		TopHooks:    topHooks,
		TopKeywords: topKeywords,
		TopHashtags: topHashtags,
		VideoCount:  len(videos),
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

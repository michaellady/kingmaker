// Package prompt generates OpusClip-compatible prompts from analyzed video patterns.
package prompt

import (
	"fmt"
	"strings"

	"github.com/mikelady/kingmaker/internal/analyzer"
	"github.com/mikelady/kingmaker/internal/hooks"
	"github.com/mikelady/kingmaker/internal/keywords"
)

// Options configures prompt generation behavior.
type Options struct {
	MaxPrompts      int    // Maximum number of prompts to generate (default 5)
	MaxPromptLength int    // Maximum length per prompt in characters (default 280)
	Query           string // Original search query for context
}

// DefaultOptions returns sensible defaults for prompt generation.
func DefaultOptions() Options {
	return Options{
		MaxPrompts:      5,
		MaxPromptLength: 280,
	}
}

// Generate creates OpusClip-compatible prompts from analyzed patterns.
// Prompts are designed for ClipAnything's natural language format:
// Subject + Action + Setting + Emotion/Sentiment
func Generate(patterns analyzer.Patterns, opts Options) []string {
	if patterns.VideoCount == 0 && len(patterns.TopKeywords) == 0 {
		return []string{}
	}

	// Apply defaults
	if opts.MaxPrompts <= 0 {
		opts.MaxPrompts = 5
	}
	if opts.MaxPromptLength <= 0 {
		opts.MaxPromptLength = 280
	}

	var prompts []string

	// Extract key elements
	topKeywords := extractTopWords(patterns.TopKeywords, 5)
	topHashtags := extractTopTags(patterns.TopHashtags, 3)
	hookTypes := categorizeHooks(patterns.TopHooks)

	// Generate prompts based on patterns found

	// 1. Keyword-focused prompt
	if len(topKeywords) > 0 {
		prompt := generateKeywordPrompt(topKeywords, opts.Query)
		if prompt != "" {
			prompts = append(prompts, truncate(prompt, opts.MaxPromptLength))
		}
	}

	// 2. Hook-based prompts (one per hook type found)
	for hookType, examples := range hookTypes {
		if len(prompts) >= opts.MaxPrompts {
			break
		}
		prompt := generateHookPrompt(hookType, examples, topKeywords)
		if prompt != "" {
			prompts = append(prompts, truncate(prompt, opts.MaxPromptLength))
		}
	}

	// 3. Hashtag/trend-focused prompt
	if len(topHashtags) > 0 && len(prompts) < opts.MaxPrompts {
		prompt := generateTrendPrompt(topHashtags, topKeywords)
		if prompt != "" {
			prompts = append(prompts, truncate(prompt, opts.MaxPromptLength))
		}
	}

	// 4. Engagement-focused prompt
	if len(prompts) < opts.MaxPrompts && len(topKeywords) > 0 {
		prompt := generateEngagementPrompt(topKeywords, patterns.VideoCount)
		if prompt != "" {
			prompts = append(prompts, truncate(prompt, opts.MaxPromptLength))
		}
	}

	// Limit to max prompts
	if len(prompts) > opts.MaxPrompts {
		prompts = prompts[:opts.MaxPrompts]
	}

	return prompts
}

func extractTopWords(kws []keywords.Keyword, n int) []string {
	result := make([]string, 0, n)
	for i, kw := range kws {
		if i >= n {
			break
		}
		result = append(result, kw.Word)
	}
	return result
}

func extractTopTags(tags []analyzer.Hashtag, n int) []string {
	result := make([]string, 0, n)
	for i, tag := range tags {
		if i >= n {
			break
		}
		result = append(result, tag.Tag)
	}
	return result
}

func categorizeHooks(allHooks []hooks.Hook) map[hooks.HookType][]string {
	result := make(map[hooks.HookType][]string)
	for _, h := range allHooks {
		if h.Frequency > 0 {
			result[h.Type] = append(result[h.Type], h.Pattern)
		}
	}
	return result
}

func generateKeywordPrompt(keywords []string, query string) string {
	if len(keywords) == 0 {
		return ""
	}

	kwList := strings.Join(keywords, ", ")

	if query != "" {
		return fmt.Sprintf("Find clips about %s featuring discussions of %s with high energy moments", query, kwList)
	}
	return fmt.Sprintf("Find engaging moments where the creator discusses %s with enthusiasm or excitement", kwList)
}

func generateHookPrompt(hookType hooks.HookType, patterns []string, keywords []string) string {
	kwContext := ""
	if len(keywords) > 0 {
		kwContext = " about " + strings.Join(keywords[:min(2, len(keywords))], " or ")
	}

	switch hookType {
	case hooks.Question:
		return fmt.Sprintf("Clip moments where the creator asks thought-provoking questions%s and provides surprising answers", kwContext)
	case hooks.Numerical:
		return fmt.Sprintf("Find segments with numbered tips, lists, or step-by-step explanations%s", kwContext)
	case hooks.PowerWord:
		return fmt.Sprintf("Extract high-impact moments with bold claims or revelations%s", kwContext)
	case hooks.CuriosityGap:
		return fmt.Sprintf("Find teaser moments that create suspense or curiosity%s before revealing insights", kwContext)
	default:
		return ""
	}
}

func generateTrendPrompt(hashtags []string, keywords []string) string {
	if len(hashtags) == 0 {
		return ""
	}

	trendTerms := strings.Join(hashtags, ", ")

	if len(keywords) > 0 {
		return fmt.Sprintf("Find viral-worthy moments discussing trending topics like %s with keywords %s", trendTerms, strings.Join(keywords[:min(3, len(keywords))], ", "))
	}
	return fmt.Sprintf("Extract shareable clips covering trending topics: %s", trendTerms)
}

func generateEngagementPrompt(keywords []string, videoCount int) string {
	if len(keywords) == 0 {
		return ""
	}

	return fmt.Sprintf("Find the most engaging moments with clear value delivery about %s - reactions, demonstrations, or aha moments", strings.Join(keywords[:min(3, len(keywords))], ", "))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Truncate at word boundary if possible
	truncated := s[:maxLen-3]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLen/2 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Package metadataprompt generates OpusClip "create-default" prompts for viral video metadata.
// Unlike the prompt package (which creates prompts for finding clips), this package
// creates prompts that help generate viral-worthy titles and descriptions.
package metadataprompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/mikelady/kingmaker/internal/analyzer"
)

// OpenAIClient defines the interface for LLM completion.
type OpenAIClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
	TokensUsed() int64
}

// MetadataPromptGenerator defines the interface for generating metadata prompts.
type MetadataPromptGenerator interface {
	Generate(ctx context.Context, patterns analyzer.Patterns, opts Options) (string, error)
}

// Options configures the prompt generation.
type Options struct {
	Niche     string // Content niche (e.g., "AI vibe coding")
	MaxLength int    // Maximum prompt length (0 = no limit)
}

// Generator creates metadata prompts using LLM.
type Generator struct {
	client OpenAIClient
}

// NewGenerator creates a new metadata prompt generator.
func NewGenerator(client OpenAIClient) *Generator {
	return &Generator{client: client}
}

// Generate creates an OpusClip create-default prompt based on analyzed patterns.
func (g *Generator) Generate(ctx context.Context, patterns analyzer.Patterns, opts Options) (string, error) {
	// Build the system prompt for the LLM
	systemPrompt := buildSystemPrompt(patterns, opts)

	// Call OpenAI to synthesize the prompt
	result, err := g.client.Complete(ctx, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	// Trim and validate result
	result = strings.TrimSpace(result)
	if opts.MaxLength > 0 && len(result) > opts.MaxLength {
		result = result[:opts.MaxLength]
	}

	return result, nil
}

// buildSystemPrompt creates the prompt for the LLM to generate metadata instructions.
func buildSystemPrompt(patterns analyzer.Patterns, opts Options) string {
	var sb strings.Builder

	sb.WriteString("You are an expert at creating viral YouTube Shorts content. ")
	sb.WriteString("Based on the following analysis of top-performing videos, ")
	sb.WriteString("create an OpusClip 'create-default' prompt that will generate ")
	sb.WriteString("viral-worthy titles and descriptions.\n\n")

	// Add niche context
	niche := opts.Niche
	if niche == "" {
		niche = "viral content"
	}
	sb.WriteString(fmt.Sprintf("Niche: %s\n", niche))
	sb.WriteString(fmt.Sprintf("Videos analyzed: %d\n\n", patterns.VideoCount))

	// Add hooks analysis
	if len(patterns.TopHooks) > 0 {
		sb.WriteString("Top performing hooks:\n")
		for i, h := range patterns.TopHooks {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s (%s): used %d times\n", h.Pattern, h.Type, h.Frequency))
		}
		sb.WriteString("\n")
	}

	// Add keywords analysis
	if len(patterns.TopKeywords) > 0 {
		sb.WriteString("Top keywords:\n")
		for i, kw := range patterns.TopKeywords {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s (frequency: %d)\n", kw.Word, kw.Frequency))
		}
		sb.WriteString("\n")
	}

	// Add hashtags analysis
	if len(patterns.TopHashtags) > 0 {
		sb.WriteString("Top hashtags:\n")
		for i, ht := range patterns.TopHashtags {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("- #%s (used %d times)\n", ht.Tag, ht.Frequency))
		}
		sb.WriteString("\n")
	}

	// Add title metrics
	if patterns.TitleMetrics.AvgLength > 0 {
		sb.WriteString("Title metrics:\n")
		sb.WriteString(fmt.Sprintf("- Average length: %d characters\n", patterns.TitleMetrics.AvgLength))
		sb.WriteString(fmt.Sprintf("- Average word count: %d words\n", patterns.TitleMetrics.AvgWords))
		sb.WriteString(fmt.Sprintf("- Hook density: %.0f%% of titles use hooks\n", patterns.TitleMetrics.HookDensity*100))

		if len(patterns.TitleMetrics.CommonPatterns) > 0 {
			sb.WriteString("- Common title patterns:\n")
			for _, p := range patterns.TitleMetrics.CommonPatterns {
				sb.WriteString(fmt.Sprintf("  - '%s' (used %d times)\n", p.Name, p.Count))
			}
		}
		sb.WriteString("\n")
	}

	// Request format
	sb.WriteString("Create a single, focused prompt (2-4 sentences) that instructs OpusClip how to:\n")
	sb.WriteString("1. Generate attention-grabbing titles using the proven hooks and patterns above\n")
	sb.WriteString("2. Write compelling descriptions with relevant keywords and hashtags\n")
	sb.WriteString("3. Match the style and energy of successful videos in this niche\n\n")
	sb.WriteString("The prompt should be actionable and specific to this niche. ")
	sb.WriteString("Do not include any explanations, just output the prompt itself.")

	return sb.String()
}

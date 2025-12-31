package prompt

import (
	"strings"
	"testing"

	"github.com/mikelady/kingmaker/internal/analyzer"
	"github.com/mikelady/kingmaker/internal/hooks"
	"github.com/mikelady/kingmaker/internal/keywords"
)

func TestGenerate_EmptyPatterns(t *testing.T) {
	patterns := analyzer.Patterns{}
	prompts := Generate(patterns, DefaultOptions())

	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts for empty patterns, got %d", len(prompts))
	}
}

func TestGenerate_WithHooksAndKeywords(t *testing.T) {
	patterns := analyzer.Patterns{
		TopHooks: []hooks.Hook{
			{Type: hooks.Question, Pattern: "how", Frequency: 5},
			{Type: hooks.Numerical, Pattern: "numerical", Frequency: 3},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "ai", Frequency: 10, Score: 0.1},
			{Word: "coding", Frequency: 8, Score: 0.08},
			{Word: "cursor", Frequency: 5, Score: 0.05},
		},
		TopHashtags: []analyzer.Hashtag{
			{Tag: "programming", Frequency: 7},
			{Tag: "tech", Frequency: 5},
		},
		VideoCount: 20,
	}

	prompts := Generate(patterns, DefaultOptions())

	if len(prompts) == 0 {
		t.Fatal("expected at least one prompt")
	}

	// Check prompts contain relevant content
	allPromptsText := strings.Join(prompts, " ")

	// Should reference top keywords
	if !strings.Contains(strings.ToLower(allPromptsText), "ai") {
		t.Error("expected prompts to reference 'ai' keyword")
	}
}

func TestGenerate_RespectsMaxPrompts(t *testing.T) {
	patterns := analyzer.Patterns{
		TopHooks: []hooks.Hook{
			{Type: hooks.Question, Pattern: "how", Frequency: 5},
			{Type: hooks.Question, Pattern: "why", Frequency: 4},
			{Type: hooks.Numerical, Pattern: "numerical", Frequency: 3},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "ai", Frequency: 10},
			{Word: "coding", Frequency: 8},
		},
		VideoCount: 20,
	}

	opts := Options{MaxPrompts: 2}
	prompts := Generate(patterns, opts)

	if len(prompts) > 2 {
		t.Errorf("expected max 2 prompts, got %d", len(prompts))
	}
}

func TestGenerate_PromptLength(t *testing.T) {
	patterns := analyzer.Patterns{
		TopHooks: []hooks.Hook{
			{Type: hooks.Question, Pattern: "how", Frequency: 10},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "artificial", Frequency: 10},
			{Word: "intelligence", Frequency: 9},
			{Word: "programming", Frequency: 8},
			{Word: "development", Frequency: 7},
		},
		VideoCount: 50,
	}

	opts := Options{MaxPromptLength: 100}
	prompts := Generate(patterns, opts)

	for i, p := range prompts {
		if len(p) > 100 {
			t.Errorf("prompt %d exceeds max length: %d > 100", i, len(p))
		}
	}
}

func TestGenerate_IncludesHashtags(t *testing.T) {
	patterns := analyzer.Patterns{
		TopHashtags: []analyzer.Hashtag{
			{Tag: "viralshorts", Frequency: 15},
			{Tag: "trending", Frequency: 10},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "content", Frequency: 5},
		},
		VideoCount: 10,
	}

	prompts := Generate(patterns, DefaultOptions())

	// At least one prompt should mention hashtags or trending topics
	found := false
	for _, p := range prompts {
		lower := strings.ToLower(p)
		if strings.Contains(lower, "viral") || strings.Contains(lower, "trending") {
			found = true
			break
		}
	}
	if !found && len(patterns.TopHashtags) > 0 {
		t.Error("expected prompts to incorporate hashtag trends")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.MaxPrompts <= 0 {
		t.Error("MaxPrompts should be positive")
	}
	if opts.MaxPromptLength <= 0 {
		t.Error("MaxPromptLength should be positive")
	}
}

func TestOptions_Defaults(t *testing.T) {
	// Zero values should use defaults
	opts := Options{}
	patterns := analyzer.Patterns{
		TopKeywords: []keywords.Keyword{{Word: "test", Frequency: 1}},
		VideoCount:  1,
	}

	// Should not panic with zero options
	prompts := Generate(patterns, opts)
	_ = prompts
}

func TestPromptFormat(t *testing.T) {
	patterns := analyzer.Patterns{
		TopHooks: []hooks.Hook{
			{Type: hooks.Question, Pattern: "how", Frequency: 5},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "cursor", Frequency: 10},
			{Word: "claude", Frequency: 8},
		},
		VideoCount: 15,
	}

	prompts := Generate(patterns, DefaultOptions())

	// Prompts should be actionable for ClipAnything
	for _, p := range prompts {
		// Should not be empty
		if strings.TrimSpace(p) == "" {
			t.Error("prompt should not be empty")
		}
		// Note: prompts may start with lowercase (e.g., action verbs) - this is acceptable
	}
}

func TestGenerateWithQuery(t *testing.T) {
	patterns := analyzer.Patterns{
		TopKeywords: []keywords.Keyword{
			{Word: "vibe", Frequency: 10},
			{Word: "coding", Frequency: 8},
		},
		VideoCount: 10,
	}

	opts := Options{
		Query:      "AI vibe coding",
		MaxPrompts: 3,
	}

	prompts := Generate(patterns, opts)

	// Should incorporate query context
	allText := strings.ToLower(strings.Join(prompts, " "))
	if !strings.Contains(allText, "vibe") && !strings.Contains(allText, "coding") {
		t.Error("expected prompts to incorporate query keywords")
	}
}

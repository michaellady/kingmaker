package metadataprompt

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mikelady/kingmaker/internal/analyzer"
	"github.com/mikelady/kingmaker/internal/hooks"
	"github.com/mikelady/kingmaker/internal/keywords"
)

// mockOpenAIClient implements openai.OpenAIClient for testing
type mockOpenAIClient struct {
	response   string
	err        error
	lastPrompt string
	callCount  int
}

func (m *mockOpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	m.callCount++
	m.lastPrompt = prompt
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockOpenAIClient) TokensUsed() int64 {
	return 0
}

func TestNewGenerator(t *testing.T) {
	mock := &mockOpenAIClient{}
	gen := NewGenerator(mock)

	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
}

func TestGenerate_Success(t *testing.T) {
	mock := &mockOpenAIClient{
		response: "Create viral Shorts about AI coding with hooks like 'I built X in Y minutes'. Focus on quick wins and impressive results.",
	}

	gen := NewGenerator(mock)

	patterns := analyzer.Patterns{
		TopHooks: []hooks.Hook{
			{Type: hooks.Question, Pattern: "how", Frequency: 5},
			{Type: hooks.Numerical, Pattern: "5 tips", Frequency: 3},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "ai", Frequency: 10, Score: 1.0},
			{Word: "coding", Frequency: 8, Score: 0.8},
		},
		TopHashtags: []analyzer.Hashtag{
			{Tag: "ai", Frequency: 7},
			{Tag: "coding", Frequency: 5},
		},
		TitleMetrics: analyzer.TitleMetrics{
			AvgLength:   45,
			AvgWords:    8,
			HookDensity: 0.75,
		},
		VideoCount: 20,
	}

	opts := Options{
		Niche: "AI vibe coding",
	}

	result, err := gen.Generate(context.Background(), patterns, opts)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", mock.callCount)
	}
}

func TestGenerate_EmptyPatterns(t *testing.T) {
	mock := &mockOpenAIClient{
		response: "Generic viral content prompt",
	}

	gen := NewGenerator(mock)
	patterns := analyzer.Patterns{}
	opts := Options{Niche: "tech"}

	result, err := gen.Generate(context.Background(), patterns, opts)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result even with empty patterns")
	}
}

func TestGenerate_APIError(t *testing.T) {
	mock := &mockOpenAIClient{
		err: errors.New("API error"),
	}

	gen := NewGenerator(mock)
	patterns := analyzer.Patterns{VideoCount: 5}
	opts := Options{Niche: "tech"}

	_, err := gen.Generate(context.Background(), patterns, opts)

	if err == nil {
		t.Error("expected error from API")
	}
}

func TestGenerate_IncludesPatternInfo(t *testing.T) {
	mock := &mockOpenAIClient{
		response: "Generated prompt",
	}

	gen := NewGenerator(mock)

	patterns := analyzer.Patterns{
		TopHooks: []hooks.Hook{
			{Type: hooks.Question, Pattern: "how", Frequency: 10},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "cursor", Frequency: 15, Score: 1.0},
		},
		TopHashtags: []analyzer.Hashtag{
			{Tag: "vibecodingAI", Frequency: 8},
		},
		TitleMetrics: analyzer.TitleMetrics{
			AvgLength:   50,
			HookDensity: 0.8,
		},
		VideoCount: 25,
	}

	opts := Options{Niche: "cursor AI coding"}

	_, err := gen.Generate(context.Background(), patterns, opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify the prompt includes key pattern info
	prompt := mock.lastPrompt

	if !strings.Contains(prompt, "cursor AI coding") {
		t.Error("prompt should include niche")
	}
	if !strings.Contains(prompt, "25") {
		t.Error("prompt should include video count")
	}
	if !strings.Contains(prompt, "cursor") {
		t.Error("prompt should include top keyword")
	}
	if !strings.Contains(prompt, "how") {
		t.Error("prompt should include top hook")
	}
}

func TestGenerate_DefaultNiche(t *testing.T) {
	mock := &mockOpenAIClient{
		response: "Generated prompt",
	}

	gen := NewGenerator(mock)
	patterns := analyzer.Patterns{VideoCount: 10}
	opts := Options{} // No niche specified

	_, err := gen.Generate(context.Background(), patterns, opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should use default niche
	if !strings.Contains(mock.lastPrompt, "viral") {
		t.Error("prompt should include viral content context")
	}
}

func TestOptions(t *testing.T) {
	opts := Options{
		Niche:     "AI coding",
		MaxLength: 500,
	}

	if opts.Niche != "AI coding" {
		t.Errorf("Niche = %s, want 'AI coding'", opts.Niche)
	}
	if opts.MaxLength != 500 {
		t.Errorf("MaxLength = %d, want 500", opts.MaxLength)
	}
}

func TestGenerator_Interface(t *testing.T) {
	// Verify Generator implements MetadataPromptGenerator interface
	var _ MetadataPromptGenerator = (*Generator)(nil)
}

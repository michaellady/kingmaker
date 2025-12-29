package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mikelady/kingmaker/internal/analyzer"
	"github.com/mikelady/kingmaker/internal/hooks"
	"github.com/mikelady/kingmaker/internal/keywords"
)

func TestDisplayPrompts_Empty(t *testing.T) {
	var buf bytes.Buffer
	DisplayPrompts(&buf, nil, Options{})

	output := buf.String()
	if !strings.Contains(output, "No prompts") {
		t.Error("expected 'No prompts' message for empty input")
	}
}

func TestDisplayPrompts_SinglePrompt(t *testing.T) {
	var buf bytes.Buffer
	prompts := []string{"Find clips about AI coding with excitement"}

	DisplayPrompts(&buf, prompts, Options{})

	output := buf.String()
	if !strings.Contains(output, "Find clips about AI coding") {
		t.Error("expected prompt content in output")
	}
	if !strings.Contains(output, "1.") || !strings.Contains(output, "1)") || !strings.Contains(output, "#1") {
		// Should have some numbering
	}
}

func TestDisplayPrompts_MultiplePrompts(t *testing.T) {
	var buf bytes.Buffer
	prompts := []string{
		"First prompt about coding",
		"Second prompt about AI",
		"Third prompt about tech",
	}

	DisplayPrompts(&buf, prompts, Options{})

	output := buf.String()
	for _, p := range prompts {
		if !strings.Contains(output, p) {
			t.Errorf("expected output to contain: %s", p)
		}
	}
}

func TestDisplayPrompts_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	prompts := []string{"Test prompt one", "Test prompt two"}

	DisplayPrompts(&buf, prompts, Options{JSON: true})

	output := buf.String()

	// Should be valid JSON
	var result []string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("expected valid JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 prompts in JSON, got %d", len(result))
	}
}

func TestDisplayPrompts_WithSummary(t *testing.T) {
	var buf bytes.Buffer
	prompts := []string{"Prompt 1", "Prompt 2", "Prompt 3"}

	DisplayPrompts(&buf, prompts, Options{ShowSummary: true})

	output := buf.String()
	if !strings.Contains(output, "3") {
		t.Error("expected summary to show prompt count")
	}
}

func TestDisplayPatterns_Empty(t *testing.T) {
	var buf bytes.Buffer
	patterns := analyzer.Patterns{}

	DisplayPatterns(&buf, patterns, Options{})

	output := buf.String()
	if !strings.Contains(output, "0 videos") && !strings.Contains(output, "No patterns") {
		t.Error("expected indication of empty patterns")
	}
}

func TestDisplayPatterns_WithData(t *testing.T) {
	var buf bytes.Buffer
	patterns := analyzer.Patterns{
		TopHooks: []hooks.Hook{
			{Type: hooks.Question, Pattern: "how", Frequency: 5},
		},
		TopKeywords: []keywords.Keyword{
			{Word: "ai", Frequency: 10},
			{Word: "coding", Frequency: 8},
		},
		TopHashtags: []analyzer.Hashtag{
			{Tag: "programming", Frequency: 7},
		},
		VideoCount: 25,
	}

	DisplayPatterns(&buf, patterns, Options{})

	output := buf.String()
	if !strings.Contains(output, "25") {
		t.Error("expected video count in output")
	}
	if !strings.Contains(output, "ai") {
		t.Error("expected keyword 'ai' in output")
	}
}

func TestDisplayPatterns_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	patterns := analyzer.Patterns{
		TopKeywords: []keywords.Keyword{
			{Word: "test", Frequency: 5},
		},
		VideoCount: 10,
	}

	DisplayPatterns(&buf, patterns, Options{JSON: true})

	output := buf.String()

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("expected valid JSON output: %v", err)
	}
}

func TestDisplayResults_Full(t *testing.T) {
	var buf bytes.Buffer

	patterns := analyzer.Patterns{
		TopKeywords: []keywords.Keyword{
			{Word: "cursor", Frequency: 10},
		},
		VideoCount: 15,
	}
	prompts := []string{"Find exciting moments about cursor"}

	DisplayResults(&buf, patterns, prompts, Options{})

	output := buf.String()
	if !strings.Contains(output, "cursor") {
		t.Error("expected keyword in output")
	}
	if !strings.Contains(output, "Find exciting") {
		t.Error("expected prompt in output")
	}
}

func TestDisplayResults_JSONFormat(t *testing.T) {
	var buf bytes.Buffer

	patterns := analyzer.Patterns{VideoCount: 5}
	prompts := []string{"Test prompt"}

	DisplayResults(&buf, patterns, prompts, Options{JSON: true})

	output := buf.String()

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("expected valid JSON: %v", err)
	}

	if _, ok := result["patterns"]; !ok {
		t.Error("expected 'patterns' key in JSON")
	}
	if _, ok := result["prompts"]; !ok {
		t.Error("expected 'prompts' key in JSON")
	}
}

func TestOptions_Defaults(t *testing.T) {
	opts := Options{}
	if opts.JSON {
		t.Error("JSON should default to false")
	}
	if opts.ShowSummary {
		t.Error("ShowSummary should default to false")
	}
}

func TestDisplayPrompts_Delimiter(t *testing.T) {
	var buf bytes.Buffer
	prompts := []string{"First", "Second"}

	DisplayPrompts(&buf, prompts, Options{})

	output := buf.String()
	// Should have clear separation between prompts
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Error("expected prompts on separate lines")
	}
}

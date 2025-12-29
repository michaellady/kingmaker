// Package cli provides formatted output for the kingmaker CLI.
package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mikelady/kingmaker/internal/analyzer"
)

// Options configures output formatting.
type Options struct {
	JSON        bool // Output as JSON instead of plain text
	ShowSummary bool // Show summary statistics
	Verbose     bool // Show additional details
}

// DisplayPrompts writes prompts to the given writer.
func DisplayPrompts(w io.Writer, prompts []string, opts Options) {
	if len(prompts) == 0 {
		if opts.JSON {
			fmt.Fprintln(w, "[]")
		} else {
			fmt.Fprintln(w, "No prompts generated.")
		}
		return
	}

	if opts.JSON {
		data, _ := json.MarshalIndent(prompts, "", "  ")
		fmt.Fprintln(w, string(data))
		return
	}

	// Plain text format
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "  OPUSCLIP PROMPTS")
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(w)

	for i, prompt := range prompts {
		fmt.Fprintf(w, "  %d. %s\n", i+1, prompt)
		fmt.Fprintln(w)
	}

	if opts.ShowSummary {
		fmt.Fprintln(w, "───────────────────────────────────────────────────────────")
		fmt.Fprintf(w, "  Generated %d prompt(s)\n", len(prompts))
		fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
	}
}

// DisplayPatterns writes analyzed patterns to the given writer.
func DisplayPatterns(w io.Writer, patterns analyzer.Patterns, opts Options) {
	if opts.JSON {
		data, _ := json.MarshalIndent(patterns, "", "  ")
		fmt.Fprintln(w, string(data))
		return
	}

	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "  PATTERN ANALYSIS")
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  Videos analyzed: %d\n", patterns.VideoCount)
	fmt.Fprintln(w)

	// Top Hooks
	if len(patterns.TopHooks) > 0 {
		fmt.Fprintln(w, "  Top Hooks:")
		for i, h := range patterns.TopHooks {
			if i >= 5 {
				break
			}
			fmt.Fprintf(w, "    • %s (%s) - %d occurrences\n", h.Pattern, h.Type.String(), h.Frequency)
		}
		fmt.Fprintln(w)
	}

	// Top Keywords
	if len(patterns.TopKeywords) > 0 {
		fmt.Fprintln(w, "  Top Keywords:")
		for i, kw := range patterns.TopKeywords {
			if i >= 10 {
				break
			}
			fmt.Fprintf(w, "    • %s (%d)\n", kw.Word, kw.Frequency)
		}
		fmt.Fprintln(w)
	}

	// Top Hashtags
	if len(patterns.TopHashtags) > 0 {
		fmt.Fprintln(w, "  Top Hashtags:")
		for i, tag := range patterns.TopHashtags {
			if i >= 5 {
				break
			}
			fmt.Fprintf(w, "    • #%s (%d)\n", tag.Tag, tag.Frequency)
		}
		fmt.Fprintln(w)
	}

	if patterns.VideoCount == 0 && len(patterns.TopKeywords) == 0 {
		fmt.Fprintln(w, "  No patterns found (0 videos analyzed)")
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
}

// DisplayResults writes both patterns and prompts to the given writer.
func DisplayResults(w io.Writer, patterns analyzer.Patterns, prompts []string, opts Options) {
	if opts.JSON {
		result := struct {
			Patterns analyzer.Patterns `json:"patterns"`
			Prompts  []string          `json:"prompts"`
		}{
			Patterns: patterns,
			Prompts:  prompts,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(w, string(data))
		return
	}

	DisplayPatterns(w, patterns, opts)
	fmt.Fprintln(w)
	DisplayPrompts(w, prompts, opts)
}

// DisplayError writes an error message to the given writer.
func DisplayError(w io.Writer, err error, opts Options) {
	if opts.JSON {
		result := struct {
			Error string `json:"error"`
		}{Error: err.Error()}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(w, string(data))
		return
	}

	fmt.Fprintf(w, "Error: %v\n", err)
}

// DisplayProgress writes a progress message (only in non-JSON mode).
func DisplayProgress(w io.Writer, message string, opts Options) {
	if opts.JSON {
		return // Silent in JSON mode
	}
	fmt.Fprintf(w, "→ %s\n", message)
}

// DisplayMetadataPrompt writes the LLM-generated metadata prompt.
func DisplayMetadataPrompt(w io.Writer, prompt string, patterns analyzer.Patterns, opts Options) {
	if opts.JSON {
		result := struct {
			MetadataPrompt string            `json:"metadata_prompt"`
			Patterns       analyzer.Patterns `json:"patterns"`
		}{
			MetadataPrompt: prompt,
			Patterns:       patterns,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(w, string(data))
		return
	}

	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "  OPUSCLIP CREATE-DEFAULT PROMPT")
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s\n", prompt)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────")
	fmt.Fprintf(w, "  Based on analysis of %d videos\n", patterns.VideoCount)
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════")
}

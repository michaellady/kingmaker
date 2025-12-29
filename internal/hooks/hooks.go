// Package hooks provides extraction of engagement hooks from video titles.
// Hooks are patterns that attract viewer attention (questions, numbers, power words).
package hooks

import (
	"regexp"
	"sort"
	"strings"
)

// HookType represents the category of an engagement hook.
type HookType int

const (
	Question HookType = iota
	Numerical
	PowerWord
	CuriosityGap
)

// String returns the string representation of a HookType.
func (h HookType) String() string {
	switch h {
	case Question:
		return "Question"
	case Numerical:
		return "Numerical"
	case PowerWord:
		return "PowerWord"
	case CuriosityGap:
		return "CuriosityGap"
	default:
		return "Unknown"
	}
}

// Hook represents an extracted engagement pattern from video titles.
type Hook struct {
	Type      HookType
	Pattern   string   // The matched pattern (e.g., "how", "5", "secret")
	Frequency int      // How many times this pattern appeared
	Examples  []string // Example titles containing this hook (up to 3)
}

// Regex patterns for different hook types
var (
	questionWords = []string{"what", "how", "why", "who", "when", "where", "which", "can", "do", "does", "is", "are", "will", "should"}
	numericalRe   = regexp.MustCompile(`(?i)\b(\d+)\s*(ways?|tips?|tricks?|secrets?|reasons?|things?|steps?|mistakes?|hacks?|ideas?|methods?|strategies?|rules?|facts?|signs?|lessons?)\b`)
	topNumericalRe = regexp.MustCompile(`(?i)\btop\s*(\d+)\b`)

	curiosityPatterns = []string{
		`(?i)you won'?t believe`,
		`(?i)this is why`,
		`(?i)here'?s what`,
		`(?i)the reason`,
		`(?i)nobody tells you`,
		`(?i)what happened`,
		`(?i)what they don'?t`,
		`(?i)the truth about`,
		`(?i)you need to know`,
		`(?i)stop doing this`,
	}
)

var powerWords = []string{
	"secret", "secrets",
	"shocking", "shocked",
	"amazing", "amazed",
	"ultimate",
	"insane", "crazy",
	"unbelievable", "incredible",
	"powerful", "proven",
	"instant", "instantly",
	"free", "guaranteed",
	"exclusive", "limited",
	"urgent", "warning",
	"banned", "hidden",
	"revealed", "exposed",
	"game-changer", "life-changing",
	"mind-blowing", "jaw-dropping",
	"breakthrough", "revolutionary",
}

// GetPowerWords returns the list of power words used for hook detection.
func GetPowerWords() []string {
	result := make([]string, len(powerWords))
	copy(result, powerWords)
	return result
}

// ExtractHooks analyzes titles and returns detected engagement hooks.
// Results are sorted by frequency (highest first) within each type.
func ExtractHooks(titles []string) []Hook {
	if len(titles) == 0 {
		return []Hook{}
	}

	// Track patterns and their occurrences
	questionCounts := make(map[string][]string)
	numericalCounts := make(map[string][]string)
	powerWordCounts := make(map[string][]string)
	curiosityCounts := make(map[string][]string)

	for _, title := range titles {
		lower := strings.ToLower(title)

		// Check for question words at the start of title
		for _, qw := range questionWords {
			if matchesQuestionPattern(lower, qw) {
				questionCounts[qw] = appendExample(questionCounts[qw], title)
			}
		}

		// Check for numerical patterns
		if matches := numericalRe.FindStringSubmatch(lower); len(matches) > 0 {
			key := "numerical"
			numericalCounts[key] = appendExample(numericalCounts[key], title)
		}
		if matches := topNumericalRe.FindStringSubmatch(lower); len(matches) > 0 {
			key := "top-n"
			numericalCounts[key] = appendExample(numericalCounts[key], title)
		}

		// Check for power words
		for _, pw := range powerWords {
			if strings.Contains(lower, pw) {
				numericalCounts[pw] = nil // just for detection, actual tracking below
				powerWordCounts[pw] = appendExample(powerWordCounts[pw], title)
			}
		}

		// Check for curiosity gap patterns
		for i, pattern := range curiosityPatterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(lower) {
				key := curiosityPatternKey(i)
				curiosityCounts[key] = appendExample(curiosityCounts[key], title)
			}
		}
	}

	// Build result slice
	var hooks []Hook

	// Add question hooks
	for pattern, examples := range questionCounts {
		hooks = append(hooks, Hook{
			Type:      Question,
			Pattern:   pattern,
			Frequency: len(examples),
			Examples:  limitExamples(examples, 3),
		})
	}

	// Add numerical hooks
	for pattern, examples := range numericalCounts {
		if pattern == "numerical" || pattern == "top-n" {
			hooks = append(hooks, Hook{
				Type:      Numerical,
				Pattern:   pattern,
				Frequency: len(examples),
				Examples:  limitExamples(examples, 3),
			})
		}
	}

	// Add power word hooks
	for pattern, examples := range powerWordCounts {
		hooks = append(hooks, Hook{
			Type:      PowerWord,
			Pattern:   pattern,
			Frequency: len(examples),
			Examples:  limitExamples(examples, 3),
		})
	}

	// Add curiosity gap hooks
	for pattern, examples := range curiosityCounts {
		hooks = append(hooks, Hook{
			Type:      CuriosityGap,
			Pattern:   pattern,
			Frequency: len(examples),
			Examples:  limitExamples(examples, 3),
		})
	}

	// Sort by frequency descending within each type
	sort.Slice(hooks, func(i, j int) bool {
		if hooks[i].Type != hooks[j].Type {
			return hooks[i].Type < hooks[j].Type
		}
		return hooks[i].Frequency > hooks[j].Frequency
	})

	return hooks
}

// matchesQuestionPattern checks if text starts with or contains a question word pattern.
func matchesQuestionPattern(text, word string) bool {
	// Check if starts with the question word
	if strings.HasPrefix(text, word+" ") || strings.HasPrefix(text, word+"'") {
		return true
	}
	// Also match question words after common prefixes
	prefixes := []string{"- ", "| ", ": "}
	for _, prefix := range prefixes {
		if strings.Contains(text, prefix+word+" ") {
			return true
		}
	}
	return false
}

func appendExample(examples []string, title string) []string {
	return append(examples, title)
}

func limitExamples(examples []string, max int) []string {
	if len(examples) <= max {
		return examples
	}
	return examples[:max]
}

func curiosityPatternKey(index int) string {
	keys := []string{
		"won't believe",
		"this is why",
		"here's what",
		"the reason",
		"nobody tells",
		"what happened",
		"what they don't",
		"the truth about",
		"need to know",
		"stop doing",
	}
	if index < len(keys) {
		return keys[index]
	}
	return "curiosity"
}

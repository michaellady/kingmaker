// Package keywords provides keyword extraction with term frequency scoring.
package keywords

import (
	"sort"

	"github.com/mikelady/kingmaker/internal/text"
)

// Keyword represents an extracted keyword with its frequency and score.
type Keyword struct {
	Word      string
	Frequency int
	Score     float64
}

// ExtractKeywords extracts the top N keywords from a collection of texts.
// Keywords are ranked by term frequency with stop words removed.
// Returns keywords sorted by frequency (highest first).
func ExtractKeywords(texts []string, topN int) []Keyword {
	if len(texts) == 0 || topN <= 0 {
		return []Keyword{}
	}

	// Count word frequencies across all texts
	wordCounts := make(map[string]int)
	totalWords := 0

	for _, t := range texts {
		tokens := text.Tokenize(t)
		filtered := text.RemoveStopWords(tokens)

		for _, word := range filtered {
			// Skip very short words (likely noise)
			if len(word) < 2 {
				continue
			}
			wordCounts[word]++
			totalWords++
		}
	}

	if totalWords == 0 {
		return []Keyword{}
	}

	// Convert to slice for sorting
	keywords := make([]Keyword, 0, len(wordCounts))
	for word, count := range wordCounts {
		// Calculate TF score (term frequency)
		score := float64(count) / float64(totalWords)
		keywords = append(keywords, Keyword{
			Word:      word,
			Frequency: count,
			Score:     score,
		})
	}

	// Sort by frequency descending
	sort.Slice(keywords, func(i, j int) bool {
		if keywords[i].Frequency != keywords[j].Frequency {
			return keywords[i].Frequency > keywords[j].Frequency
		}
		// Tie-breaker: alphabetical order
		return keywords[i].Word < keywords[j].Word
	})

	// Return top N
	if len(keywords) > topN {
		keywords = keywords[:topN]
	}

	return keywords
}

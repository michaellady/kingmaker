package text

import (
	"regexp"
	"strings"
)

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true,
	"did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "must": true, "shall": true,
	"i": true, "me": true, "my": true, "myself": true, "we": true,
	"our": true, "ours": true, "ourselves": true, "you": true, "your": true,
	"yours": true, "yourself": true, "yourselves": true, "he": true, "him": true,
	"his": true, "himself": true, "she": true, "her": true, "hers": true,
	"herself": true, "it": true, "its": true, "itself": true, "they": true,
	"them": true, "their": true, "theirs": true, "themselves": true,
	"what": true, "which": true, "who": true, "whom": true, "this": true,
	"that": true, "these": true, "those": true, "am": true,
	"and": true, "but": true, "if": true, "or": true, "because": true,
	"as": true, "until": true, "while": true, "of": true, "at": true,
	"by": true, "for": true, "with": true, "about": true, "against": true,
	"between": true, "into": true, "through": true, "during": true,
	"before": true, "after": true, "above": true, "below": true, "to": true,
	"from": true, "up": true, "down": true, "in": true, "out": true,
	"on": true, "off": true, "over": true, "under": true, "again": true,
	"further": true, "then": true, "once": true, "here": true, "there": true,
	"when": true, "where": true, "why": true, "how": true, "all": true,
	"each": true, "few": true, "more": true, "most": true, "other": true,
	"some": true, "such": true, "no": true, "nor": true, "not": true,
	"only": true, "own": true, "same": true, "so": true, "than": true,
	"too": true, "very": true, "s": true, "t": true, "can": true,
	"just": true, "don": true, "now": true,
}

var wordRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)
var hashtagRegex = regexp.MustCompile(`#([a-zA-Z0-9]+)`)
var whitespaceRegex = regexp.MustCompile(`\s+`)

// Tokenize splits text into lowercase tokens, removing punctuation.
func Tokenize(text string) []string {
	if text == "" {
		return []string{}
	}

	matches := wordRegex.FindAllString(text, -1)
	if matches == nil {
		return []string{}
	}

	tokens := make([]string, len(matches))
	for i, match := range matches {
		tokens[i] = strings.ToLower(match)
	}
	return tokens
}

// RemoveStopWords filters out common stop words from a token slice.
func RemoveStopWords(tokens []string) []string {
	if len(tokens) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		lower := strings.ToLower(token)
		if !stopWords[lower] {
			result = append(result, lower)
		}
	}
	return result
}

// ExtractHashtags finds all hashtags in text, returning the tag without the # symbol.
func ExtractHashtags(text string) []string {
	if text == "" {
		return []string{}
	}

	matches := hashtagRegex.FindAllStringSubmatch(text, -1)
	if matches == nil {
		return []string{}
	}

	tags := make([]string, len(matches))
	for i, match := range matches {
		tags[i] = match[1] // capture group without #
	}
	return tags
}

// NGrams generates n-grams from a slice of tokens.
func NGrams(tokens []string, n int) []string {
	if len(tokens) < n || n < 1 {
		return []string{}
	}

	ngrams := make([]string, 0, len(tokens)-n+1)
	for i := 0; i <= len(tokens)-n; i++ {
		ngram := strings.Join(tokens[i:i+n], " ")
		ngrams = append(ngrams, ngram)
	}
	return ngrams
}

// NormalizeText lowercases, trims, and collapses whitespace.
func NormalizeText(text string) string {
	text = strings.ToLower(text)
	text = whitespaceRegex.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	return text
}

package text

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple sentence", "hello world", []string{"hello", "world"}},
		{"with punctuation", "Hello, World!", []string{"hello", "world"}},
		{"multiple spaces", "hello   world", []string{"hello", "world"}},
		{"empty string", "", []string{}},
		{"numbers mixed", "top 10 tips", []string{"top", "10", "tips"}},
		{"special chars", "AI-powered coding", []string{"ai", "powered", "coding"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tokenize(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tokenize(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRemoveStopWords(t *testing.T) {
	tests := []struct {
		name   string
		tokens []string
		want   []string
	}{
		{"basic removal", []string{"the", "quick", "brown", "fox"}, []string{"quick", "brown", "fox"}},
		{"all stop words", []string{"the", "a", "an", "is"}, []string{}},
		{"no stop words", []string{"quick", "brown", "fox"}, []string{"quick", "brown", "fox"}},
		{"empty input", []string{}, []string{}},
		{"mixed case handled", []string{"The", "quick"}, []string{"quick"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveStopWords(tt.tokens)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveStopWords(%v) = %v, want %v", tt.tokens, got, tt.want)
			}
		})
	}
}

func TestExtractHashtags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single hashtag", "Check out #coding", []string{"coding"}},
		{"multiple hashtags", "#AI #coding #shorts", []string{"AI", "coding", "shorts"}},
		{"no hashtags", "Hello world", []string{}},
		{"hashtag at start", "#trending video", []string{"trending"}},
		{"adjacent hashtags", "#tech#tips", []string{"tech", "tips"}},
		{"empty string", "", []string{}},
		{"hashtag with numbers", "#top10 tips", []string{"top10"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractHashtags(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractHashtags(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNGrams(t *testing.T) {
	tests := []struct {
		name   string
		tokens []string
		n      int
		want   []string
	}{
		{"bigrams", []string{"the", "quick", "brown", "fox"}, 2, []string{"the quick", "quick brown", "brown fox"}},
		{"trigrams", []string{"the", "quick", "brown", "fox"}, 3, []string{"the quick brown", "quick brown fox"}},
		{"unigrams", []string{"hello", "world"}, 1, []string{"hello", "world"}},
		{"n larger than tokens", []string{"hello"}, 2, []string{}},
		{"empty tokens", []string{}, 2, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NGrams(tt.tokens, tt.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NGrams(%v, %d) = %v, want %v", tt.tokens, tt.n, got, tt.want)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "HELLO WORLD", "hello world"},
		{"trim whitespace", "  hello  ", "hello"},
		{"multiple spaces", "hello   world", "hello world"},
		{"mixed case", "HeLLo WoRLD", "hello world"},
		{"empty string", "", ""},
		{"newlines and tabs", "hello\n\tworld", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeText(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

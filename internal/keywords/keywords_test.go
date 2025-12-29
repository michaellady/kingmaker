package keywords

import (
	"testing"
)

func TestExtractKeywords_EmptyInput(t *testing.T) {
	keywords := ExtractKeywords(nil, 10)
	if len(keywords) != 0 {
		t.Errorf("ExtractKeywords(nil) = %d keywords, want 0", len(keywords))
	}

	keywords = ExtractKeywords([]string{}, 10)
	if len(keywords) != 0 {
		t.Errorf("ExtractKeywords([]) = %d keywords, want 0", len(keywords))
	}
}

func TestExtractKeywords_SingleText(t *testing.T) {
	texts := []string{"Learn the golang programming basics"}
	keywords := ExtractKeywords(texts, 10)

	if len(keywords) == 0 {
		t.Fatal("ExtractKeywords() returned no keywords")
	}

	// Should find "golang" and "programming" but not stop words like "the"
	foundGolang := false
	foundProgramming := false
	for _, kw := range keywords {
		if kw.Word == "golang" {
			foundGolang = true
		}
		if kw.Word == "programming" {
			foundProgramming = true
		}
		// Stop words should be filtered
		if kw.Word == "the" {
			t.Error("Stop word 'the' should be filtered")
		}
	}

	if !foundGolang {
		t.Error("Expected to find 'golang'")
	}
	if !foundProgramming {
		t.Error("Expected to find 'programming'")
	}
}

func TestExtractKeywords_FrequencyCounting(t *testing.T) {
	texts := []string{
		"golang golang golang",
		"programming programming",
		"code",
	}

	keywords := ExtractKeywords(texts, 10)

	var golangKw, programmingKw, codeKw *Keyword
	for i := range keywords {
		switch keywords[i].Word {
		case "golang":
			golangKw = &keywords[i]
		case "programming":
			programmingKw = &keywords[i]
		case "code":
			codeKw = &keywords[i]
		}
	}

	if golangKw == nil || golangKw.Frequency != 3 {
		t.Errorf("golang frequency = %v, want 3", golangKw)
	}
	if programmingKw == nil || programmingKw.Frequency != 2 {
		t.Errorf("programming frequency = %v, want 2", programmingKw)
	}
	if codeKw == nil || codeKw.Frequency != 1 {
		t.Errorf("code frequency = %v, want 1", codeKw)
	}
}

func TestExtractKeywords_TopN(t *testing.T) {
	texts := []string{
		"apple apple apple apple",
		"banana banana banana",
		"cherry cherry",
		"date",
	}

	keywords := ExtractKeywords(texts, 2)

	if len(keywords) != 2 {
		t.Errorf("ExtractKeywords(topN=2) returned %d, want 2", len(keywords))
	}

	// Top 2 should be apple and banana
	if keywords[0].Word != "apple" {
		t.Errorf("keywords[0] = %q, want 'apple'", keywords[0].Word)
	}
	if keywords[1].Word != "banana" {
		t.Errorf("keywords[1] = %q, want 'banana'", keywords[1].Word)
	}
}

func TestExtractKeywords_SortedByFrequency(t *testing.T) {
	texts := []string{
		"coding coding coding",
		"development development",
		"software",
	}

	keywords := ExtractKeywords(texts, 10)

	for i := 1; i < len(keywords); i++ {
		if keywords[i].Frequency > keywords[i-1].Frequency {
			t.Errorf("Keywords not sorted by frequency: [%d]=%d > [%d]=%d",
				i, keywords[i].Frequency, i-1, keywords[i-1].Frequency)
		}
	}
}

func TestExtractKeywords_StopWordsRemoved(t *testing.T) {
	texts := []string{
		"the quick brown fox jumps over the lazy dog",
		"a an the is are was were be been being",
	}

	keywords := ExtractKeywords(texts, 100)

	stopWords := []string{"the", "a", "an", "is", "are", "was", "were", "be", "been", "being", "over"}
	for _, kw := range keywords {
		for _, sw := range stopWords {
			if kw.Word == sw {
				t.Errorf("Stop word %q should be filtered", sw)
			}
		}
	}
}

func TestExtractKeywords_CaseInsensitive(t *testing.T) {
	texts := []string{
		"Golang GOLANG golang GoLang",
	}

	keywords := ExtractKeywords(texts, 10)

	if len(keywords) != 1 {
		t.Fatalf("Expected 1 keyword, got %d", len(keywords))
	}

	if keywords[0].Word != "golang" {
		t.Errorf("Word = %q, want 'golang' (lowercase)", keywords[0].Word)
	}
	if keywords[0].Frequency != 4 {
		t.Errorf("Frequency = %d, want 4", keywords[0].Frequency)
	}
}

func TestExtractKeywords_Score(t *testing.T) {
	texts := []string{
		"keyword keyword keyword",
		"another another",
		"single",
	}

	keywords := ExtractKeywords(texts, 10)

	// Score should be calculated (TF for MVP)
	for _, kw := range keywords {
		if kw.Score <= 0 {
			t.Errorf("Keyword %q has score %f, want > 0", kw.Word, kw.Score)
		}
	}

	// Higher frequency should have higher score
	if len(keywords) >= 2 {
		if keywords[0].Score < keywords[1].Score {
			t.Error("Higher frequency keyword should have higher score")
		}
	}
}

func TestExtractKeywords_TopNZero(t *testing.T) {
	texts := []string{"test data"}
	keywords := ExtractKeywords(texts, 0)

	if len(keywords) != 0 {
		t.Errorf("ExtractKeywords(topN=0) = %d, want 0", len(keywords))
	}
}

func TestExtractKeywords_TopNNegative(t *testing.T) {
	texts := []string{"test data"}
	keywords := ExtractKeywords(texts, -1)

	if len(keywords) != 0 {
		t.Errorf("ExtractKeywords(topN=-1) = %d, want 0", len(keywords))
	}
}

func TestExtractKeywords_TopNLargerThanResults(t *testing.T) {
	texts := []string{"apple banana"}
	keywords := ExtractKeywords(texts, 100)

	// Should return all available keywords, not 100
	if len(keywords) != 2 {
		t.Errorf("ExtractKeywords(topN=100) = %d, want 2", len(keywords))
	}
}

func TestExtractKeywords_MultipleTexts(t *testing.T) {
	texts := []string{
		"learn programming basics",
		"programming is fun",
		"advanced programming techniques",
	}

	keywords := ExtractKeywords(texts, 10)

	var progKw *Keyword
	for i := range keywords {
		if keywords[i].Word == "programming" {
			progKw = &keywords[i]
			break
		}
	}

	if progKw == nil {
		t.Fatal("Expected to find 'programming'")
	}
	if progKw.Frequency != 3 {
		t.Errorf("'programming' frequency = %d, want 3", progKw.Frequency)
	}
}

func TestExtractKeywords_EmptyStringsIgnored(t *testing.T) {
	texts := []string{"", "valid keyword", ""}
	keywords := ExtractKeywords(texts, 10)

	if len(keywords) == 0 {
		t.Error("ExtractKeywords() should handle empty strings gracefully")
	}
}

func TestExtractKeywords_MinWordLength(t *testing.T) {
	texts := []string{"a go is to do programming"}
	keywords := ExtractKeywords(texts, 10)

	// Single letter words and very short words (that are also stop words) should be filtered
	for _, kw := range keywords {
		if len(kw.Word) < 2 {
			t.Errorf("Single-char word %q should be filtered", kw.Word)
		}
	}
}

func TestKeyword_Fields(t *testing.T) {
	kw := Keyword{
		Word:      "testing",
		Frequency: 5,
		Score:     0.25,
	}

	if kw.Word != "testing" {
		t.Errorf("Word = %q, want 'testing'", kw.Word)
	}
	if kw.Frequency != 5 {
		t.Errorf("Frequency = %d, want 5", kw.Frequency)
	}
	if kw.Score != 0.25 {
		t.Errorf("Score = %f, want 0.25", kw.Score)
	}
}

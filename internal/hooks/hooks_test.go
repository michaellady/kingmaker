package hooks

import (
	"reflect"
	"testing"
)

func TestHookType_String(t *testing.T) {
	tests := []struct {
		hookType HookType
		want     string
	}{
		{Question, "Question"},
		{Numerical, "Numerical"},
		{PowerWord, "PowerWord"},
		{CuriosityGap, "CuriosityGap"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.hookType.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractHooks_EmptyInput(t *testing.T) {
	hooks := ExtractHooks(nil)
	if len(hooks) != 0 {
		t.Errorf("ExtractHooks(nil) = %d hooks, want 0", len(hooks))
	}

	hooks = ExtractHooks([]string{})
	if len(hooks) != 0 {
		t.Errorf("ExtractHooks([]) = %d hooks, want 0", len(hooks))
	}
}

func TestExtractHooks_QuestionPatterns(t *testing.T) {
	titles := []string{
		"What is the best way to learn Go?",
		"How to cook pasta perfectly",
		"Why do cats purr?",
		"Who invented the telephone?",
		"When should you use goroutines?",
		"Where to find cheap flights",
	}

	hooks := ExtractHooks(titles)

	// Should find question hooks
	var questionHooks []Hook
	for _, h := range hooks {
		if h.Type == Question {
			questionHooks = append(questionHooks, h)
		}
	}

	if len(questionHooks) == 0 {
		t.Error("ExtractHooks() found no question hooks")
	}

	// Verify frequency is tracked
	foundWhat := false
	for _, h := range questionHooks {
		if h.Pattern == "what" {
			foundWhat = true
			if h.Frequency != 1 {
				t.Errorf("'what' frequency = %d, want 1", h.Frequency)
			}
		}
	}
	if !foundWhat {
		t.Error("Did not find 'what' pattern")
	}
}

func TestExtractHooks_NumericalPatterns(t *testing.T) {
	titles := []string{
		"5 ways to improve your code",
		"10 tips for better sleep",
		"3 tricks every developer should know",
		"7 secrets to success",
		"Top 15 mistakes beginners make",
	}

	hooks := ExtractHooks(titles)

	var numericalHooks []Hook
	for _, h := range hooks {
		if h.Type == Numerical {
			numericalHooks = append(numericalHooks, h)
		}
	}

	if len(numericalHooks) == 0 {
		t.Error("ExtractHooks() found no numerical hooks")
	}
}

func TestExtractHooks_PowerWordPatterns(t *testing.T) {
	titles := []string{
		"The SECRET to getting rich",
		"SHOCKING truth about diets",
		"This AMAZING trick will change your life",
		"The ULTIMATE guide to productivity",
		"INSANE results in 30 days",
	}

	hooks := ExtractHooks(titles)

	var powerWordHooks []Hook
	for _, h := range hooks {
		if h.Type == PowerWord {
			powerWordHooks = append(powerWordHooks, h)
		}
	}

	if len(powerWordHooks) == 0 {
		t.Error("ExtractHooks() found no power word hooks")
	}

	// Check that specific power words are found
	foundSecret := false
	for _, h := range powerWordHooks {
		if h.Pattern == "secret" {
			foundSecret = true
		}
	}
	if !foundSecret {
		t.Error("Did not find 'secret' power word")
	}
}

func TestExtractHooks_CuriosityGapPatterns(t *testing.T) {
	titles := []string{
		"You won't believe what happened next",
		"This is why you're always tired",
		"Here's what nobody tells you about",
		"The reason you're not losing weight",
	}

	hooks := ExtractHooks(titles)

	var curiosityHooks []Hook
	for _, h := range hooks {
		if h.Type == CuriosityGap {
			curiosityHooks = append(curiosityHooks, h)
		}
	}

	if len(curiosityHooks) == 0 {
		t.Error("ExtractHooks() found no curiosity gap hooks")
	}
}

func TestExtractHooks_FrequencyTracking(t *testing.T) {
	titles := []string{
		"How to do X",
		"How to do Y",
		"How to do Z",
		"What is A",
		"What is B",
	}

	hooks := ExtractHooks(titles)

	howCount := 0
	whatCount := 0
	for _, h := range hooks {
		if h.Type == Question {
			if h.Pattern == "how" {
				howCount = h.Frequency
			}
			if h.Pattern == "what" {
				whatCount = h.Frequency
			}
		}
	}

	if howCount != 3 {
		t.Errorf("'how' frequency = %d, want 3", howCount)
	}
	if whatCount != 2 {
		t.Errorf("'what' frequency = %d, want 2", whatCount)
	}
}

func TestExtractHooks_CaseInsensitive(t *testing.T) {
	titles := []string{
		"HOW to do this",
		"how TO do that",
		"How To Do Another",
	}

	hooks := ExtractHooks(titles)

	found := false
	for _, h := range hooks {
		if h.Type == Question && h.Pattern == "how" {
			found = true
			if h.Frequency != 3 {
				t.Errorf("'how' frequency = %d, want 3", h.Frequency)
			}
		}
	}
	if !found {
		t.Error("Did not find 'how' question pattern")
	}
}

func TestExtractHooks_MixedTypes(t *testing.T) {
	titles := []string{
		"5 SECRET ways to save money",
		"How I made $10000 in a week",
	}

	hooks := ExtractHooks(titles)

	types := make(map[HookType]bool)
	for _, h := range hooks {
		types[h.Type] = true
	}

	if !types[Numerical] {
		t.Error("Expected Numerical hook type")
	}
	if !types[PowerWord] {
		t.Error("Expected PowerWord hook type")
	}
	if !types[Question] {
		t.Error("Expected Question hook type")
	}
}

func TestExtractHooks_SortedByFrequency(t *testing.T) {
	titles := []string{
		"How to A", "How to B", "How to C", // 3x how
		"What is X", // 1x what
		"Why is Y", "Why is Z", // 2x why
	}

	hooks := ExtractHooks(titles)

	// Filter to question hooks only
	var questionHooks []Hook
	for _, h := range hooks {
		if h.Type == Question {
			questionHooks = append(questionHooks, h)
		}
	}

	if len(questionHooks) < 2 {
		t.Fatal("Expected at least 2 question hooks")
	}

	// Should be sorted by frequency descending
	for i := 1; i < len(questionHooks); i++ {
		if questionHooks[i].Frequency > questionHooks[i-1].Frequency {
			t.Errorf("Hooks not sorted by frequency: %v", questionHooks)
		}
	}
}

func TestHook_Fields(t *testing.T) {
	h := Hook{
		Type:      PowerWord,
		Pattern:   "secret",
		Frequency: 5,
		Examples:  []string{"The secret to success"},
	}

	if h.Type != PowerWord {
		t.Errorf("Type = %v, want PowerWord", h.Type)
	}
	if h.Pattern != "secret" {
		t.Errorf("Pattern = %q, want %q", h.Pattern, "secret")
	}
	if h.Frequency != 5 {
		t.Errorf("Frequency = %d, want 5", h.Frequency)
	}
	if !reflect.DeepEqual(h.Examples, []string{"The secret to success"}) {
		t.Errorf("Examples = %v, want [The secret to success]", h.Examples)
	}
}

func TestGetPowerWords_ReturnsNonEmpty(t *testing.T) {
	words := GetPowerWords()
	if len(words) == 0 {
		t.Error("GetPowerWords() returned empty slice")
	}

	// Check some expected power words
	wordSet := make(map[string]bool)
	for _, w := range words {
		wordSet[w] = true
	}

	expected := []string{"secret", "amazing", "shocking", "ultimate", "insane"}
	for _, exp := range expected {
		if !wordSet[exp] {
			t.Errorf("Power words missing %q", exp)
		}
	}
}

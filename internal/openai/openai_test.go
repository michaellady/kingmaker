package openai

import (
	"context"
	"errors"
	"testing"
)

// mockOpenAIService implements OpenAIService for testing
type mockOpenAIService struct {
	response   string
	err        error
	callCount  int
	lastPrompt string
	lastModel  string
}

func (m *mockOpenAIService) CreateChatCompletion(ctx context.Context, model, prompt string) (string, int, error) {
	m.callCount++
	m.lastPrompt = prompt
	m.lastModel = model
	if m.err != nil {
		return "", 0, m.err
	}
	// Simulate token usage: roughly 4 chars per token for prompt + response
	tokens := (len(prompt) + len(m.response)) / 4
	if tokens == 0 {
		tokens = 1
	}
	return m.response, tokens, nil
}

func TestNewClient(t *testing.T) {
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
	if client.TokensUsed() != 0 {
		t.Errorf("Initial tokens should be 0, got %d", client.TokensUsed())
	}
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Error("expected error for empty API key")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	client, err := NewClient("test-api-key", WithModel("gpt-4"))
	if err != nil {
		t.Fatalf("NewClient with options failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestComplete_Success(t *testing.T) {
	mock := &mockOpenAIService{
		response: "This is a test response",
	}

	client := &Client{
		service: mock,
		model:   DefaultModel,
	}

	ctx := context.Background()
	result, err := client.Complete(ctx, "Test prompt")

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if result != "This is a test response" {
		t.Errorf("expected 'This is a test response', got '%s'", result)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", mock.callCount)
	}
	if mock.lastPrompt != "Test prompt" {
		t.Errorf("expected prompt 'Test prompt', got '%s'", mock.lastPrompt)
	}
	if mock.lastModel != DefaultModel {
		t.Errorf("expected model '%s', got '%s'", DefaultModel, mock.lastModel)
	}
}

func TestComplete_EmptyPrompt(t *testing.T) {
	client := &Client{
		service: &mockOpenAIService{},
		model:   DefaultModel,
	}

	_, err := client.Complete(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty prompt")
	}
}

func TestComplete_APIError(t *testing.T) {
	mock := &mockOpenAIService{
		err: errors.New("API error"),
	}

	client := &Client{
		service: mock,
		model:   DefaultModel,
	}

	_, err := client.Complete(context.Background(), "Test prompt")
	if err == nil {
		t.Error("expected error from API")
	}
}

func TestComplete_TracksTokens(t *testing.T) {
	mock := &mockOpenAIService{
		response: "Response",
	}

	client := &Client{
		service: mock,
		model:   DefaultModel,
	}

	ctx := context.Background()

	// First call
	_, err := client.Complete(ctx, "First prompt")
	if err != nil {
		t.Fatalf("First Complete failed: %v", err)
	}
	firstTokens := client.TokensUsed()
	if firstTokens == 0 {
		t.Error("TokensUsed should be > 0 after first call")
	}

	// Second call
	_, err = client.Complete(ctx, "Second prompt")
	if err != nil {
		t.Fatalf("Second Complete failed: %v", err)
	}
	secondTokens := client.TokensUsed()
	if secondTokens <= firstTokens {
		t.Errorf("TokensUsed should increase: first=%d, second=%d", firstTokens, secondTokens)
	}
}

func TestComplete_UsesConfiguredModel(t *testing.T) {
	mock := &mockOpenAIService{
		response: "Response",
	}

	client := &Client{
		service: mock,
		model:   "gpt-4",
	}

	_, err := client.Complete(context.Background(), "Test")
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if mock.lastModel != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", mock.lastModel)
	}
}

func TestClient_Interface(t *testing.T) {
	// Verify Client implements OpenAIClient interface
	var _ OpenAIClient = (*Client)(nil)
}

func TestDefaultModel(t *testing.T) {
	// Verify default model is gpt-4o-mini for cost efficiency
	if DefaultModel != "gpt-4o-mini" {
		t.Errorf("DefaultModel should be 'gpt-4o-mini', got '%s'", DefaultModel)
	}
}

func TestWithModel(t *testing.T) {
	opts := &clientOptions{}
	WithModel("gpt-4")(opts)
	if opts.model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", opts.model)
	}
}

func TestTokensUsed_ThreadSafe(t *testing.T) {
	mock := &mockOpenAIService{
		response: "Response",
	}

	client := &Client{
		service: mock,
		model:   DefaultModel,
	}

	// Run concurrent calls
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = client.Complete(context.Background(), "Test prompt")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// TokensUsed should be consistent
	tokens := client.TokensUsed()
	if tokens == 0 {
		t.Error("TokensUsed should be > 0 after concurrent calls")
	}
}

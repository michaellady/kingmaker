// Package openai provides a client for OpenAI's ChatCompletion API.
package openai

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/sashabaranov/go-openai"
)

// DefaultModel is the default model for cost efficiency.
const DefaultModel = "gpt-4o-mini"

// OpenAIService abstracts the OpenAI API for testing.
type OpenAIService interface {
	CreateChatCompletion(ctx context.Context, model, prompt string) (string, int, error)
}

// OpenAIClient is the interface for the OpenAI client.
type OpenAIClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
	TokensUsed() int64
}

// Client wraps OpenAI API calls with token tracking.
type Client struct {
	service    OpenAIService
	model      string
	tokensUsed int64
}

// clientOptions holds optional configuration for the client.
type clientOptions struct {
	model string
}

// ClientOption is a function that configures the client.
type ClientOption func(*clientOptions)

// WithModel sets a custom model for the client.
func WithModel(model string) ClientOption {
	return func(o *clientOptions) {
		o.model = model
	}
}

// realOpenAIService wraps the actual OpenAI API client.
type realOpenAIService struct {
	client *openai.Client
}

func (s *realOpenAIService) CreateChatCompletion(ctx context.Context, model, prompt string) (string, int, error) {
	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", 0, err
	}

	if len(resp.Choices) == 0 {
		return "", 0, errors.New("no response choices returned from OpenAI API")
	}

	totalTokens := resp.Usage.TotalTokens
	return resp.Choices[0].Message.Content, totalTokens, nil
}

// NewClient creates a new OpenAI client with the given API key.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("API key is required")
	}

	options := &clientOptions{
		model: DefaultModel,
	}
	for _, opt := range opts {
		opt(options)
	}

	openaiClient := openai.NewClient(apiKey)
	service := &realOpenAIService{client: openaiClient}

	return &Client{
		service: service,
		model:   options.model,
	}, nil
}

// Complete sends a prompt to the OpenAI API and returns the response.
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	response, tokens, err := c.service.CreateChatCompletion(ctx, c.model, prompt)
	if err != nil {
		return "", err
	}

	atomic.AddInt64(&c.tokensUsed, int64(tokens))
	return response, nil
}

// TokensUsed returns the total tokens used across all API calls.
func (c *Client) TokensUsed() int64 {
	return atomic.LoadInt64(&c.tokensUsed)
}

package config

import (
	"errors"
	"os"
)

// Config holds application configuration.
type Config struct {
	YouTubeAPIKey string
	OpenAIAPIKey  string // Optional, required for metadata mode
	MaxResults    int
	HTTPTimeout   int // seconds
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("YOUTUBE_API_KEY environment variable is required")
	}

	return &Config{
		YouTubeAPIKey: apiKey,
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"), // Optional
		MaxResults:    50,
		HTTPTimeout:   30,
	}, nil
}

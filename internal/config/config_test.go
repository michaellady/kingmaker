package config

import (
	"os"
	"testing"
)

func TestLoadConfig_FromEnv(t *testing.T) {
	os.Setenv("YOUTUBE_API_KEY", "test-api-key")
	defer os.Unsetenv("YOUTUBE_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.YouTubeAPIKey != "test-api-key" {
		t.Errorf("YouTubeAPIKey = %q, want %q", cfg.YouTubeAPIKey, "test-api-key")
	}
}

func TestLoadConfig_MissingAPIKey(t *testing.T) {
	os.Unsetenv("YOUTUBE_API_KEY")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for missing API key, got nil")
	}
}

func TestConfig_Defaults(t *testing.T) {
	os.Setenv("YOUTUBE_API_KEY", "test-key")
	defer os.Unsetenv("YOUTUBE_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxResults != 50 {
		t.Errorf("MaxResults = %d, want %d", cfg.MaxResults, 50)
	}

	if cfg.HTTPTimeout != 30 {
		t.Errorf("HTTPTimeout = %d, want %d", cfg.HTTPTimeout, 30)
	}
}

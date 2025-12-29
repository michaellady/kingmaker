// Package shorts provides YouTube Shorts detection via URL redirect checking.
// YouTube redirects /shorts/{id} URLs to /watch?v={id} for non-Shorts videos.
package shorts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/mikelady/kingmaker/internal/httpclient"
)

const youtubeBaseURL = "https://www.youtube.com/shorts/"

// ShortsChecker defines the interface for checking if videos are Shorts.
type ShortsChecker interface {
	// IsShort checks if a single video ID is a YouTube Short.
	// Returns true if the video is a Short, false otherwise.
	IsShort(ctx context.Context, videoID string) (bool, error)

	// CheckBatch checks multiple video IDs concurrently.
	// Returns a map of videoID -> isShort, and an error if any checks failed.
	CheckBatch(ctx context.Context, videoIDs []string) (map[string]bool, error)
}

// Checker implements ShortsChecker using HTTP HEAD requests.
type Checker struct {
	client httpclient.HTTPClient
}

// NewChecker creates a new Shorts checker with the given HTTP client.
// The client should NOT follow redirects (use httpclient.NewNoRedirectClient).
func NewChecker(client httpclient.HTTPClient) *Checker {
	return &Checker{client: client}
}

// IsShort checks if a video ID corresponds to a YouTube Short.
// It makes a HEAD request to youtube.com/shorts/{id}:
// - HTTP 200 = video is a Short
// - HTTP 3xx (redirect) = video is NOT a Short (redirects to /watch?v=)
func (c *Checker) IsShort(ctx context.Context, videoID string) (bool, error) {
	if videoID == "" {
		return false, errors.New("video ID cannot be empty")
	}

	url := shortsURL(videoID)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	// 200 OK means it's a Short
	// 3xx redirects mean it's not a Short (redirects to /watch?v=)
	return resp.StatusCode == http.StatusOK, nil
}

// CheckBatch checks multiple video IDs concurrently.
// Returns results for all successfully checked videos.
// If any checks fail, returns partial results along with an error.
func (c *Checker) CheckBatch(ctx context.Context, videoIDs []string) (map[string]bool, error) {
	if len(videoIDs) == 0 {
		return make(map[string]bool), nil
	}

	results := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []error
	var errMu sync.Mutex

	for _, id := range videoIDs {
		wg.Add(1)
		go func(videoID string) {
			defer wg.Done()

			isShort, err := c.IsShort(ctx, videoID)
			if err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", videoID, err))
				errMu.Unlock()
				return
			}

			mu.Lock()
			results[videoID] = isShort
			mu.Unlock()
		}(id)
	}

	wg.Wait()

	var combinedErr error
	if len(errs) > 0 {
		combinedErr = fmt.Errorf("failed to check %d video(s): %v", len(errs), errs[0])
	}

	return results, combinedErr
}

// shortsURL constructs the YouTube Shorts URL for a video ID.
func shortsURL(videoID string) string {
	return youtubeBaseURL + videoID
}

package shorts

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockHTTPClient implements httpclient.HTTPClient for testing
type mockHTTPClient struct {
	statusCodes map[string]int
	errors      map[string]error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	return nil, nil
}

func (m *mockHTTPClient) Head(url string) (*http.Response, error) {
	if err, ok := m.errors[url]; ok {
		return nil, err
	}
	status := http.StatusNotFound
	if code, ok := m.statusCodes[url]; ok {
		status = code
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.Head(req.URL.String())
}

func newMockClient() *mockHTTPClient {
	return &mockHTTPClient{
		statusCodes: make(map[string]int),
		errors:      make(map[string]error),
	}
}

func TestIsShort_ReturnsTrue_WhenStatus200(t *testing.T) {
	mock := newMockClient()
	mock.statusCodes["https://www.youtube.com/shorts/abc123"] = http.StatusOK

	checker := NewChecker(mock)
	isShort, err := checker.IsShort(context.Background(), "abc123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isShort {
		t.Error("expected IsShort to return true for 200 status")
	}
}

func TestIsShort_ReturnsFalse_WhenStatus303(t *testing.T) {
	mock := newMockClient()
	mock.statusCodes["https://www.youtube.com/shorts/xyz789"] = http.StatusSeeOther

	checker := NewChecker(mock)
	isShort, err := checker.IsShort(context.Background(), "xyz789")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isShort {
		t.Error("expected IsShort to return false for 303 status")
	}
}

func TestIsShort_ReturnsFalse_WhenStatus302(t *testing.T) {
	mock := newMockClient()
	mock.statusCodes["https://www.youtube.com/shorts/redirect"] = http.StatusFound // 302

	checker := NewChecker(mock)
	isShort, err := checker.IsShort(context.Background(), "redirect")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isShort {
		t.Error("expected IsShort to return false for 302 status")
	}
}

func TestIsShort_ReturnsError_WhenHTTPFails(t *testing.T) {
	mock := newMockClient()
	mock.errors["https://www.youtube.com/shorts/fail"] = context.DeadlineExceeded

	checker := NewChecker(mock)
	_, err := checker.IsShort(context.Background(), "fail")

	if err == nil {
		t.Error("expected error when HTTP request fails")
	}
}

func TestIsShort_EmptyVideoID(t *testing.T) {
	mock := newMockClient()
	checker := NewChecker(mock)

	_, err := checker.IsShort(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty video ID")
	}
}

func TestCheckBatch_ReturnsAllResults(t *testing.T) {
	mock := newMockClient()
	mock.statusCodes["https://www.youtube.com/shorts/short1"] = http.StatusOK
	mock.statusCodes["https://www.youtube.com/shorts/short2"] = http.StatusOK
	mock.statusCodes["https://www.youtube.com/shorts/notshort"] = http.StatusSeeOther

	checker := NewChecker(mock)
	results, err := checker.CheckBatch(context.Background(), []string{"short1", "short2", "notshort"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if !results["short1"] {
		t.Error("expected short1 to be a Short")
	}
	if !results["short2"] {
		t.Error("expected short2 to be a Short")
	}
	if results["notshort"] {
		t.Error("expected notshort to NOT be a Short")
	}
}

func TestCheckBatch_EmptyInput(t *testing.T) {
	mock := newMockClient()
	checker := NewChecker(mock)

	results, err := checker.CheckBatch(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestCheckBatch_PartialErrors(t *testing.T) {
	mock := newMockClient()
	mock.statusCodes["https://www.youtube.com/shorts/good"] = http.StatusOK
	mock.errors["https://www.youtube.com/shorts/bad"] = context.DeadlineExceeded

	checker := NewChecker(mock)
	results, err := checker.CheckBatch(context.Background(), []string{"good", "bad"})

	// Should return partial results with error
	if err == nil {
		t.Error("expected error for partial failure")
	}
	// Should still have the successful result
	if !results["good"] {
		t.Error("expected 'good' to be in results")
	}
}

func TestShortsURL(t *testing.T) {
	url := shortsURL("abc123")
	expected := "https://www.youtube.com/shorts/abc123"
	if url != expected {
		t.Errorf("shortsURL = %q, want %q", url, expected)
	}
}

func TestChecker_Interface(t *testing.T) {
	// Verify Checker implements ShortsChecker interface
	var _ ShortsChecker = (*Checker)(nil)
}

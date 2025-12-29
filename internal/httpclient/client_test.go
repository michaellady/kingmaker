package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient_ReturnsNonNil(t *testing.T) {
	client := NewHTTPClient(30 * time.Second)
	if client == nil {
		t.Error("NewHTTPClient() returned nil")
	}
}

func TestNewHTTPClient_SetsTimeout(t *testing.T) {
	timeout := 15 * time.Second
	client := NewHTTPClient(timeout)

	// Type assert to access underlying client
	c, ok := client.(*http.Client)
	if !ok {
		t.Fatal("NewHTTPClient() did not return *http.Client")
	}

	if c.Timeout != timeout {
		t.Errorf("Timeout = %v, want %v", c.Timeout, timeout)
	}
}

func TestNewHTTPClient_FollowsRedirects(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			redirectCount++
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	resp, err := client.Get(server.URL + "/redirect")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if redirectCount != 1 {
		t.Errorf("redirectCount = %d, want 1", redirectCount)
	}
}

func TestNewNoRedirectClient_ReturnsNonNil(t *testing.T) {
	client := NewNoRedirectClient(30 * time.Second)
	if client == nil {
		t.Error("NewNoRedirectClient() returned nil")
	}
}

func TestNewNoRedirectClient_SetsTimeout(t *testing.T) {
	timeout := 10 * time.Second
	client := NewNoRedirectClient(timeout)

	c, ok := client.(*http.Client)
	if !ok {
		t.Fatal("NewNoRedirectClient() did not return *http.Client")
	}

	if c.Timeout != timeout {
		t.Errorf("Timeout = %v, want %v", c.Timeout, timeout)
	}
}

func TestNewNoRedirectClient_DoesNotFollowRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	}))
	defer server.Close()

	client := NewNoRedirectClient(5 * time.Second)
	resp, err := client.Get(server.URL + "/redirect")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	// Should get the redirect status, not follow it
	if resp.StatusCode != http.StatusFound {
		t.Errorf("StatusCode = %d, want %d (redirect not followed)", resp.StatusCode, http.StatusFound)
	}
}

func TestNewNoRedirectClient_ReturnsRedirectLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/shorts/abc123", http.StatusFound)
	}))
	defer server.Close()

	client := NewNoRedirectClient(5 * time.Second)
	resp, err := client.Get(server.URL + "/watch?v=abc123")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location != "/shorts/abc123" {
		t.Errorf("Location = %q, want %q", location, "/shorts/abc123")
	}
}

func TestHTTPClient_Interface(t *testing.T) {
	// Verify both clients implement HTTPClient interface
	var _ HTTPClient = NewHTTPClient(time.Second)
	var _ HTTPClient = NewNoRedirectClient(time.Second)
}

func TestHTTPClient_Head(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("Method = %s, want HEAD", r.Method)
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	resp, err := client.Head(server.URL)
	if err != nil {
		t.Fatalf("Head() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHTTPClient_Do(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "test-value" {
			t.Error("Custom header not received")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	req.Header.Set("X-Custom", "test-value")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

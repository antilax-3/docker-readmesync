package github

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &Client{HTTP: server.Client(), RawURL: server.URL}
}

func TestReadme(t *testing.T) {
	t.Parallel()

	var path string

	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		_, _ = io.WriteString(w, "# readme-sync\n")
	})

	readme, err := client.Readme(t.Context(), "antilax-3/docker-readmesync", "renovate/golang-1.x")
	if err != nil || readme != "# readme-sync\n" {
		t.Fatalf("got %q, %v", readme, err)
	}

	if want := "/antilax-3/docker-readmesync/renovate/golang-1.x/README.md"; path != want {
		t.Fatalf("requested %s, want %s", path, want)
	}
}

func TestReadmeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		is      error
		message string
	}{
		{
			name:    "Missing",
			handler: func(w http.ResponseWriter, _ *http.Request) { http.NotFound(w, nil) },
			is:      ErrNotFound,
		},
		{
			name:    "GitHubFailing",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) },
			message: "503 Service Unavailable",
		},
		{
			name: "TooLarge",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, strings.Repeat("x", MaxReadmeBytes+1))
			},
			is: ErrTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := newClient(t, tt.handler).Readme(t.Context(), "antilax-3/docker-readmesync", "master")

			if tt.is != nil && !errors.Is(err, tt.is) {
				t.Fatalf("got %v, want %v", err, tt.is)
			}

			if tt.message != "" && (err == nil || !strings.Contains(err.Error(), tt.message)) {
				t.Fatalf("got %v, want an error containing %q", err, tt.message)
			}
		})
	}
}

package readmesync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/antilax-3/docker-readmesync/internal/github"
)

type source map[string]string // "<repo>@<branch>" -> README

func (s source) Readme(_ context.Context, repo, branch string) (string, error) {
	readme, ok := s[repo+"@"+branch]
	if !ok {
		return "", github.ErrNotFound
	}

	return readme, nil
}

type sink struct {
	stored map[string]string
	err    error
}

func (s *sink) SetFullDescription(_ context.Context, repo, description string) error {
	if s.err != nil {
		return s.err
	}

	s.stored[repo] = description

	return nil
}

func serve(t *testing.T, method, query string, s *sink) (int, string) {
	t.Helper()

	output := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(output) })

	syncer := &Syncer{
		Source: source{
			"antilax-3/docker-readmesync@master":              "# readme-sync\n",
			"antilax-3/docker-readmesync@renovate/golang-1.x": "# branch\n",
		},
		Sink: s,
	}

	rec := httptest.NewRecorder()
	Handler(syncer).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, "/readmesync/update?"+query, nil))

	return rec.Code, strings.TrimSpace(rec.Body.String())
}

func TestHandlerSyncs(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"MasterByDefault", "github_repo=antilax-3/docker-readmesync&dockerhub_repo=antilax3/readme-sync", "# readme-sync\n"},
		{"RequestedBranch", "github_repo=antilax-3/docker-readmesync&github_branch=renovate/golang-1.x&dockerhub_repo=antilax3/readme-sync", "# branch\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sink{stored: map[string]string{}}

			code, body := serve(t, http.MethodGet, tt.query, s)
			if code != http.StatusOK || body != "OK" {
				t.Fatalf("got %d %q, want 200 OK", code, body)
			}

			if got := s.stored["antilax3/readme-sync"]; got != tt.want {
				t.Fatalf("stored %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandlerRejects(t *testing.T) {
	tests := []struct {
		name   string
		method string
		query  string
		code   int
		body   string
	}{
		{"Post", http.MethodPost, "", http.StatusMethodNotAllowed, "Method not allowed"},
		{"MissingGitHubRepo", http.MethodGet, "dockerhub_repo=antilax3/readme-sync", http.StatusBadRequest, "Missing required fields in GET request"},
		{"MissingDockerHubRepo", http.MethodGet, "github_repo=antilax-3/docker-readmesync", http.StatusBadRequest, "Missing required fields in GET request"},
		{"GitHubRepoWithPath", http.MethodGet, "github_repo=antilax-3/docker-readmesync/x&dockerhub_repo=antilax3/readme-sync", http.StatusBadRequest, "invalid github_repo"},
		{"GitHubRepoClimbingOut", http.MethodGet, "github_repo=antilax-3/..&dockerhub_repo=antilax3/readme-sync", http.StatusBadRequest, "invalid github_repo"},
		{"BranchClimbingOut", http.MethodGet, "github_repo=antilax-3/docker-readmesync&github_branch=../../x&dockerhub_repo=antilax3/readme-sync", http.StatusBadRequest, "invalid github_branch"},
		{"DockerHubRepoWithPath", http.MethodGet, "github_repo=antilax-3/docker-readmesync&dockerhub_repo=antilax3/readme-sync/tags", http.StatusBadRequest, "invalid dockerhub_repo"},
		{"NoReadme", http.MethodGet, "github_repo=antilax-3/nope&dockerhub_repo=antilax3/readme-sync", http.StatusBadRequest, "no README.md in GitHub repository antilax-3/nope branch master"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sink{stored: map[string]string{}}

			code, body := serve(t, tt.method, tt.query, s)
			if code != tt.code || !strings.Contains(body, tt.body) {
				t.Fatalf("got %d %q, want %d containing %q", code, body, tt.code, tt.body)
			}

			if len(s.stored) != 0 {
				t.Fatalf("stored %v, want nothing written", s.stored)
			}
		})
	}
}

func TestHandlerReportsUpstreamFailures(t *testing.T) {
	s := &sink{err: fmt.Errorf("unable to log in to Docker Hub as nightah: %w", errors.New("401 Unauthorized: invalid username/password"))}

	code, body := serve(t, http.MethodGet, "github_repo=antilax-3/docker-readmesync&dockerhub_repo=antilax3/readme-sync", s)
	if code != http.StatusBadGateway || !strings.Contains(body, "invalid username/password") {
		t.Fatalf("got %d %q, want 502 carrying Docker Hub's message", code, body)
	}
}

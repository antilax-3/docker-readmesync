// Package readmesync copies a GitHub repository's README.md to a Docker Hub repository's full description, and serves
// that as an HTTP API.
package readmesync

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/antilax-3/docker-readmesync/internal/github"
)

var (
	githubRepoPattern    = regexp.MustCompile(`^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$`)
	githubBranchPattern  = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)
	dockerHubRepoPattern = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*/[a-z0-9]+(?:[._-][a-z0-9]+)*$`)
)

// ErrBadRequest marks an error the caller caused, as opposed to one GitHub or Docker Hub did.
var ErrBadRequest = errors.New("bad request")

// ReadmeSource fetches a repository's README.
type ReadmeSource interface {
	Readme(ctx context.Context, repo, branch string) (string, error)
}

// DescriptionSink stores a repository's full description.
type DescriptionSink interface {
	SetFullDescription(ctx context.Context, repo, description string) error
}

// Request is one sync: the README at GitHubRepo's GitHubBranch goes to DockerHubRepo.
type Request struct {
	GitHubRepo    string
	GitHubBranch  string
	DockerHubRepo string
}

// Validate rejects anything that is not a plain repository or branch name, so no field can reach a different path on
// GitHub or Docker Hub.
func (r Request) Validate() error {
	switch {
	case !githubRepoPattern.MatchString(r.GitHubRepo) || strings.HasSuffix(r.GitHubRepo, "/.") || strings.HasSuffix(r.GitHubRepo, "/.."):
		return fmt.Errorf("%w: invalid github_repo %q, expected <owner>/<repository>", ErrBadRequest, r.GitHubRepo)
	case !githubBranchPattern.MatchString(r.GitHubBranch) || strings.Contains(r.GitHubBranch, ".."):
		return fmt.Errorf("%w: invalid github_branch %q", ErrBadRequest, r.GitHubBranch)
	case !dockerHubRepoPattern.MatchString(r.DockerHubRepo):
		return fmt.Errorf("%w: invalid dockerhub_repo %q, expected <namespace>/<repository>", ErrBadRequest, r.DockerHubRepo)
	}

	return nil
}

// Syncer copies READMEs from Source to Sink.
type Syncer struct {
	Source ReadmeSource
	Sink   DescriptionSink
}

// Sync validates r, fetches the README and stores it as the description.
func (s *Syncer) Sync(ctx context.Context, r Request) error {
	if err := r.Validate(); err != nil {
		return err
	}

	readme, err := s.Source.Readme(ctx, r.GitHubRepo, r.GitHubBranch)

	switch {
	case errors.Is(err, github.ErrNotFound):
		return fmt.Errorf("%w: no README.md in GitHub repository %s branch %s", ErrBadRequest, r.GitHubRepo, r.GitHubBranch)
	case errors.Is(err, github.ErrTooLarge):
		return fmt.Errorf("%w: README.md in GitHub repository %s branch %s is %w", ErrBadRequest, r.GitHubRepo, r.GitHubBranch, err)
	case err != nil:
		return fmt.Errorf("unable to fetch README.md from GitHub repository %s branch %s: %w", r.GitHubRepo, r.GitHubBranch, err)
	}

	return s.Sink.SetFullDescription(ctx, r.DockerHubRepo, readme)
}

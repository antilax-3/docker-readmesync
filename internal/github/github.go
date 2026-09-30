// Package github fetches README files from GitHub repositories.
package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DefaultRawURL is where GitHub serves raw repository files.
const DefaultRawURL = "https://raw.githubusercontent.com"

// MaxReadmeBytes bounds how much of a README is read. Docker Hub caps a full description at 25,000 characters and
// answers anything longer with its own error, so the bound only stops an unexpected response being read whole.
const MaxReadmeBytes = 1 << 20

var (
	// ErrNotFound reports that the repository, the branch or the README in it does not exist.
	ErrNotFound = errors.New("not found")
	// ErrTooLarge reports a README over MaxReadmeBytes.
	ErrTooLarge = errors.New("too large")
)

// Client fetches files from GitHub.
type Client struct {
	HTTP   *http.Client
	RawURL string
}

// Readme returns the README.md at the root of repo on branch. repo is <owner>/<repository>, and a branch may contain
// slashes.
func (c *Client) Readme(ctx context.Context, repo, branch string) (string, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/README.md", c.RawURL, repo, pathEscapeEach(branch))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	switch {
	case res.StatusCode == http.StatusNotFound:
		return "", ErrNotFound
	case res.StatusCode != http.StatusOK:
		return "", fmt.Errorf("github responded %s", res.Status)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, MaxReadmeBytes+1))
	if err != nil {
		return "", err
	}

	if len(body) > MaxReadmeBytes {
		return "", fmt.Errorf("%w: over %d bytes", ErrTooLarge, MaxReadmeBytes)
	}

	return string(body), nil
}

// pathEscapeEach escapes every segment of a slash separated path, keeping the slashes, so a branch such as
// renovate/express-5.x stays two segments.
func pathEscapeEach(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}

	return strings.Join(segments, "/")
}

// Package dockerhub is a minimal client for the Docker Hub API, covering what readmesync needs: logging in and setting
// a repository's full description.
package dockerhub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultURL is the Docker Hub API root.
const DefaultURL = "https://hub.docker.com"

// maxResponseBytes bounds how much of a response is read; a repository echoes its description back, and Docker Hub
// caps that at 25,000 characters.
const maxResponseBytes = 1 << 20

// ErrNotStored reports that Docker Hub accepted a description but returned something else as the stored value.
var ErrNotStored = errors.New("docker hub accepted the description but did not store it")

// Client talks to Docker Hub as one account. Password is the account password or a personal access token; Docker Hub
// exchanges either for a bearer token.
type Client struct {
	HTTP     *http.Client
	URL      string
	Username string
	Password string
}

// ResponseError is a Docker Hub response other than 200, carrying Docker Hub's own message.
type ResponseError struct {
	Status  string
	Message string
}

func (e *ResponseError) Error() string {
	return e.Status + ": " + e.Message
}

// SetFullDescription logs in and sets the full description of repo, <namespace>/<repository>, then checks Docker Hub
// stored what was sent.
func (c *Client) SetFullDescription(ctx context.Context, repo, description string) error {
	token, err := c.token(ctx)
	if err != nil {
		return fmt.Errorf("unable to log in to Docker Hub as %s: %w", c.Username, err)
	}

	var out struct {
		FullDescription string `json:"full_description"`
	}

	in := map[string]string{"full_description": description}
	if err := c.do(ctx, http.MethodPatch, "/v2/repositories/"+repo+"/", token, in, &out); err != nil {
		return fmt.Errorf("unable to update the Docker Hub description of %s: %w", repo, err)
	}

	if out.FullDescription != description {
		return fmt.Errorf("%s: %w", repo, ErrNotStored)
	}

	return nil
}

// token exchanges the credentials for a bearer token.
func (c *Client) token(ctx context.Context) (string, error) {
	var out struct {
		AccessToken string `json:"access_token"`
	}

	in := map[string]string{"identifier": c.Username, "secret": c.Password}
	if err := c.do(ctx, http.MethodPost, "/v2/auth/token", "", in, &out); err != nil {
		return "", err
	}

	if out.AccessToken == "" {
		return "", errors.New("no access token returned")
	}

	return out.AccessToken, nil
}

// do sends in as JSON and decodes a 200 response into out. Any other status is a *ResponseError.
func (c *Client) do(ctx context.Context, method, path, token string, in, out any) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.URL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return &ResponseError{Status: res.Status, Message: message(data)}
	}

	return json.Unmarshal(data, out)
}

// message pulls the human readable message out of a Docker Hub error body, which uses either field.
func message(data []byte) string {
	var body struct {
		Detail  string `json:"detail"`
		Message string `json:"message"`
	}

	if json.Unmarshal(data, &body) == nil {
		if body.Detail != "" {
			return body.Detail
		}

		if body.Message != "" {
			return body.Message
		}
	}

	return strings.TrimSpace(string(data))
}

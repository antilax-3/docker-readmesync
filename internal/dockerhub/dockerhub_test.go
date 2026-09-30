package dockerhub

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fake is a Docker Hub that accepts one account, with either its password or its personal access token.
type fake struct {
	stored    map[string]string
	dropStore bool
}

var secrets = map[string]bool{"hunter2": true, "dckr_pat_secret": true}

func (f *fake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	_ = json.NewDecoder(r.Body).Decode(&body)

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v2/auth/token":
		if body["identifier"] != "nightah" || !secrets[body["secret"]] {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"message":"invalid username/password"}`)

			return
		}

		writeJSON(w, map[string]string{"access_token": "bearer-token"})
	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/v2/repositories/"):
		if r.Header.Get("Authorization") != "Bearer bearer-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"detail":"authentication credentials were not provided"}`)

			return
		}

		repo := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v2/repositories/"), "/")
		if !f.dropStore {
			f.stored[repo] = body["full_description"]
		}

		writeJSON(w, map[string]string{"full_description": f.stored[repo]})
	default:
		http.NotFound(w, r)
	}
}

// writeJSON encodes v as the response, answering 500 if it cannot, so a broken fake fails the test rather than
// passing on an empty body.
func writeJSON(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, _ = w.Write(data)
}

func newClient(t *testing.T, password string) (*fake, *Client) {
	t.Helper()

	f := &fake{stored: map[string]string{}}
	server := httptest.NewServer(f)
	t.Cleanup(server.Close)

	return f, &Client{HTTP: server.Client(), URL: server.URL, Username: "nightah", Password: password}
}

func TestSetFullDescription(t *testing.T) {
	t.Parallel()

	// Docker Hub exchanges an account password and a personal access token for a bearer token the same way.
	for name, password := range map[string]string{"Password": "hunter2", "PersonalAccessToken": "dckr_pat_secret"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f, client := newClient(t, password)

			if err := client.SetFullDescription(t.Context(), "antilax3/readme-sync", "# readme-sync\n"); err != nil {
				t.Fatal(err)
			}

			if got := f.stored["antilax3/readme-sync"]; got != "# readme-sync\n" {
				t.Fatalf("stored %q, want the description", got)
			}
		})
	}
}

func TestSetFullDescriptionBadCredentials(t *testing.T) {
	t.Parallel()

	f, client := newClient(t, "wrong")

	err := client.SetFullDescription(t.Context(), "antilax3/readme-sync", "# readme-sync\n")

	var responseErr *ResponseError
	if !errors.As(err, &responseErr) || responseErr.Message != "invalid username/password" {
		t.Fatalf("got %v, want Docker Hub's rejection", err)
	}

	if len(f.stored) != 0 {
		t.Fatalf("stored %v, want nothing written", f.stored)
	}
}

// The Node service reported success whenever Docker Hub's response carried no error field, whatever it stored.
func TestSetFullDescriptionNotStored(t *testing.T) {
	t.Parallel()

	f, client := newClient(t, "dckr_pat_secret")
	f.dropStore = true

	if err := client.SetFullDescription(t.Context(), "antilax3/readme-sync", "# readme-sync\n"); !errors.Is(err, ErrNotStored) {
		t.Fatalf("got %v, want ErrNotStored", err)
	}
}

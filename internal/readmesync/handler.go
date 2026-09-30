package readmesync

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

// Handler answers a GET on any path carrying github_repo and dockerhub_repo, and optionally github_branch, which
// defaults to master. It responds 200 once the description is stored, 400 for a request that can never succeed, and
// 502 when GitHub or Docker Hub fails. Each sync is logged to logger.
func Handler(syncer *Syncer, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}

		query := r.URL.Query()
		if !query.Has("github_repo") || !query.Has("dockerhub_repo") {
			http.Error(w, "Missing required fields in GET request", http.StatusBadRequest)

			return
		}

		req := Request{
			GitHubRepo:    query.Get("github_repo"),
			GitHubBranch:  query.Get("github_branch"),
			DockerHubRepo: query.Get("dockerhub_repo"),
		}

		if req.GitHubBranch == "" {
			req.GitHubBranch = "master"
		}

		if err := syncer.Sync(r.Context(), req); err != nil {
			status := http.StatusBadGateway
			if errors.Is(err, ErrBadRequest) {
				status = http.StatusBadRequest
			}

			logger.Printf("sync %s@%s to %s failed: %v", req.GitHubRepo, req.GitHubBranch, req.DockerHubRepo, err)
			http.Error(w, err.Error(), status)

			return
		}

		logger.Printf("synced %s@%s to %s", req.GitHubRepo, req.GitHubBranch, req.DockerHubRepo)
		_, _ = fmt.Fprintln(w, "OK")
	})
}

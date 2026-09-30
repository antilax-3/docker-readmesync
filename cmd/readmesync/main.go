// Command readmesync serves an API that copies a GitHub repository's README.md to a Docker Hub repository's full
// description.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/antilax-3/docker-readmesync/internal/config"
	"github.com/antilax-3/docker-readmesync/internal/dockerhub"
	"github.com/antilax-3/docker-readmesync/internal/github"
	"github.com/antilax-3/docker-readmesync/internal/readmesync"
)

const configPath = "/config/readmesync.json"

func main() {
	log.SetFlags(0)

	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	syncer := &readmesync.Syncer{
		Source: &github.Client{HTTP: client, RawURL: github.DefaultRawURL},
		Sink: &dockerhub.Client{
			HTTP:     client,
			URL:      dockerhub.DefaultURL,
			Username: cfg.DockerHubUsername,
			Password: cfg.DockerHubPassword,
		},
	}

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           readmesync.Handler(syncer),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      2 * time.Minute,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)

	go func() {
		log.Printf("Running readmesync. Listening on port %d.", cfg.Port)
		errs <- server.ListenAndServe()
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdown); err != nil {
		return err
	}

	if err := <-errs; !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

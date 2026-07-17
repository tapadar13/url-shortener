package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	defaultEndpoint = "http://127.0.0.1:8080/readyz"
	requestTimeout  = 3 * time.Second
)

func main() {
	if err := run(os.Args[1:], &http.Client{Timeout: requestTimeout}); err != nil {
		fmt.Fprintf(os.Stderr, "health check failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, client *http.Client) error {
	if len(args) > 1 {
		return errors.New("expected at most one endpoint argument")
	}
	if client == nil {
		return errors.New("HTTP client is required")
	}

	endpoint := defaultEndpoint
	if len(args) == 1 {
		endpoint = args[0]
	}

	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request readiness endpoint: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("readiness endpoint returned %s", response.Status)
	}

	return nil
}

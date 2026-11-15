package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Result stores the outcome and timing information for one target.
type Result struct {
	Target   Target
	Success  bool
	Duration time.Duration
	Attempts int
	Message  string
}

// ProbeTarget checks a target and retries failed attempts as configured.
func ProbeTarget(target Target, timeout time.Duration) Result {
	started := time.Now()
	result := Result{Target: target}
	for attempt := 1; attempt <= target.RetryCount+1; attempt++ {
		result.Attempts = attempt
		if err := probeOnce(target, timeout); err == nil {
			result.Success = true
			result.Message = "ok"
			break
		} else {
			result.Message = err.Error()
		}
	}
	result.Duration = time.Since(started)
	return result
}

func probeOnce(target Target, timeout time.Duration) error {
	switch strings.ToLower(target.Type) {
	case "http":
		return probeHTTP(target, timeout)
	case "tcp":
		return probeTCP(target, timeout)
	default:
		return fmt.Errorf("unsupported type %q", target.Type)
	}
}

func probeHTTP(target Target, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.Address, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer response.Body.Close()

	if target.ExpectedStatus != 0 && response.StatusCode != target.ExpectedStatus {
		return fmt.Errorf("unexpected status: got %d, want %d", response.StatusCode, target.ExpectedStatus)
	}
	if target.ExpectedStatus == 0 && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return fmt.Errorf("unexpected status: got %d, want a 2xx response", response.StatusCode)
	}
	if target.Contains != "" {
		// Limit the response body read so an unexpectedly large response stays bounded.
		body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}
		if !strings.Contains(string(body), target.Contains) {
			return fmt.Errorf("response body does not contain %q", target.Contains)
		}
	}
	return nil
}

func probeTCP(target Target, timeout time.Duration) error {
	connection, err := (&net.Dialer{Timeout: timeout}).Dial("tcp", target.Address)
	if err != nil {
		return fmt.Errorf("tcp connection: %w", err)
	}
	return connection.Close()
}

// RunProbes launches one goroutine per target and returns results as they finish.
func RunProbes(targets []Target, timeout time.Duration, onResult func(Result)) []Result {
	results := make(chan Result, len(targets))
	for _, target := range targets {
		// Copy the loop value so each goroutine keeps its own target.
		target := target
		go func() {
			// The assignment asks for a one-second delay to make concurrency visible.
			time.Sleep(time.Second)
			results <- ProbeTarget(target, timeout)
		}()
	}

	completed := make([]Result, 0, len(targets))
	for range targets {
		result := <-results
		completed = append(completed, result)
		if onResult != nil {
			onResult(result)
		}
	}
	return completed
}

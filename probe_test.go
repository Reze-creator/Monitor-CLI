package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestProbeHTTPStatusAndBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello Go")
	}))
	defer server.Close()

	result := ProbeTarget(Target{Name: "test", Type: "http", Address: server.URL, ExpectedStatus: 200, Contains: "Go"}, time.Second)
	if !result.Success || result.Attempts != 1 {
		t.Fatalf("expected one successful attempt, got %+v", result)
	}

	result = ProbeTarget(Target{Name: "test", Type: "http", Address: server.URL, ExpectedStatus: 201}, time.Second)
	if result.Success || !strings.Contains(result.Message, "unexpected status") {
		t.Fatalf("expected a status mismatch, got %+v", result)
	}
}

func TestProbeHTTPTimeoutRetries(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			time.Sleep(100 * time.Millisecond)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := ProbeTarget(Target{Name: "retry", Type: "http", Address: server.URL, ExpectedStatus: 200, RetryCount: 1}, 30*time.Millisecond)
	if !result.Success || result.Attempts != 2 {
		t.Fatalf("expected the second attempt to succeed, got %+v", result)
	}
}

func TestProbeTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		connection, err := listener.Accept()
		if err == nil {
			connection.Close()
		}
	}()

	result := ProbeTarget(Target{Name: "tcp", Type: "tcp", Address: listener.Addr().String()}, time.Second)
	if !result.Success {
		t.Fatalf("expected TCP connection to succeed, got %+v", result)
	}
}

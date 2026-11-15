package main

import "testing"

func TestTargetValidate(t *testing.T) {
	tests := []struct {
		name    string
		target  Target
		wantErr bool
	}{
		{"valid http", Target{Name: "web", Type: "http", Address: "https://example.com"}, false},
		{"valid tcp", Target{Name: "db", Type: "tcp", Address: "localhost:3306"}, false},
		{"unsupported type", Target{Name: "bad", Type: "udp", Address: "localhost:53"}, true},
		{"invalid retry count", Target{Name: "bad", Type: "tcp", Address: "localhost:22", RetryCount: 4}, true},
		{"invalid http url", Target{Name: "bad", Type: "http", Address: "example.com"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.target.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

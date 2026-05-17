package main

import "testing"

func TestDisplayURL(t *testing.T) {
	tests := map[string]string{
		":8080":           "http://localhost:8080",
		"127.0.0.1:18080": "http://127.0.0.1:18080",
	}
	for addr, want := range tests {
		if got := displayURL(addr); got != want {
			t.Fatalf("displayURL(%q) = %q, want %q", addr, got, want)
		}
	}
}

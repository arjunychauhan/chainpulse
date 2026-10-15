package main

import "testing"

func TestDefaultVersion(t *testing.T) {
	if version != "dev" {
		t.Fatalf("expected default version %q, got %q", "dev", version)
	}
}

package main

import "testing"

func TestListenAddress(t *testing.T) {
	t.Run("explicit address", func(t *testing.T) {
		t.Setenv("ADDR", "127.0.0.1:9000")
		t.Setenv("PORT", "10000")
		if got := listenAddress(); got != "127.0.0.1:9000" {
			t.Fatalf("listenAddress() = %q", got)
		}
	})

	t.Run("Render port", func(t *testing.T) {
		t.Setenv("ADDR", "")
		t.Setenv("PORT", "10000")
		if got := listenAddress(); got != ":10000" {
			t.Fatalf("listenAddress() = %q", got)
		}
	})

	t.Run("local default", func(t *testing.T) {
		t.Setenv("ADDR", "")
		t.Setenv("PORT", "")
		if got := listenAddress(); got != ":8080" {
			t.Fatalf("listenAddress() = %q", got)
		}
	})
}

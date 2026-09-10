package model

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRouter_Route(t *testing.T) {
	local := &MockProvider{Name: "Local"}
	external := &MockProvider{Name: "External"}
	router := NewRouter(local, external)

	ctx := context.Background()

	t.Run("sensitive goes to local", func(t *testing.T) {
		req := ModelRequest{IsSensitive: true}
		resp, err := router.Route(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Provider != "Local" {
			t.Errorf("expected Local, got %s", resp.Provider)
		}
	})

	t.Run("sensitive fail closed", func(t *testing.T) {
		failingLocal := &MockProvider{Name: "Local", ShouldFail: true}
		r := NewRouter(failingLocal, external)
		req := ModelRequest{IsSensitive: true}
		
		_, err := r.Route(ctx, req)
		if !errors.Is(err, ErrFailClosed) {
			t.Errorf("expected ErrFailClosed, got %v", err)
		}
	})

	t.Run("fallback allowed", func(t *testing.T) {
		failingLocal := &MockProvider{Name: "Local", ShouldFail: true}
		r := NewRouter(failingLocal, external)
		req := ModelRequest{FallbackAllowed: true}
		
		resp, err := r.Route(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Provider != "External" {
			t.Errorf("expected External, got %s", resp.Provider)
		}
	})

	t.Run("fallback not allowed - fail closed", func(t *testing.T) {
		failingLocal := &MockProvider{Name: "Local", ShouldFail: true}
		r := NewRouter(failingLocal, external)
		req := ModelRequest{FallbackAllowed: false}
		
		_, err := r.Route(ctx, req)
		if !errors.Is(err, ErrFailClosed) {
			t.Errorf("expected ErrFailClosed, got %v", err)
		}
	})

	t.Run("timeout behavior", func(t *testing.T) {
		// Use a short delay in the test to ensure it runs quickly, but simulates a hang
		timeoutLocal := &MockProvider{Name: "Local", Delay: 3 * time.Second}
		r := NewRouter(timeoutLocal, external)
		
		// To make the test run faster, we use a context timeout 
		// Actually, executeWithTimeout hardcodes 2 seconds. We just wait it out.
		// Wait, if it waits 2 seconds the test will take 2 seconds. That's fine for now.
		req := ModelRequest{IsSensitive: true}
		_, err := r.Route(ctx, req)
		if !errors.Is(err, ErrFailClosed) { 
			t.Errorf("expected ErrFailClosed on timeout, got %v", err)
		}
	})
}

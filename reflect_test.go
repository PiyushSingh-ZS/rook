package main

import (
	"strings"
	"testing"
)

func TestReflectionBuffer(t *testing.T) {
	entries := []struct {
		label   string
		failure string
	}{
		{"1", "FAIL: TestFoo — nil pointer dereference in store.go:42"},
		{"2", "build failed: undefined: GetAllActiveByOrg in service.go:88"},
	}

	t.Run("fresh dir returns empty/zero", func(t *testing.T) {
		fresh := t.TempDir()
		if got := reflectionAttempts(fresh); got != 0 {
			t.Fatalf("fresh reflectionAttempts = %d, want 0", got)
		}
		if got := reflectionContext(fresh); got != "" {
			t.Fatalf("fresh reflectionContext = %q, want empty", got)
		}
	})

	t.Run("accumulates and injects", func(t *testing.T) {
		wt := t.TempDir()
		for _, e := range entries {
			if err := writeReflection(wt, e.label, e.failure); err != nil {
				t.Fatalf("writeReflection(%s): %v", e.label, err)
			}
		}

		if got := reflectionAttempts(wt); got != 2 {
			t.Fatalf("reflectionAttempts = %d, want 2", got)
		}

		ctxBlock := reflectionContext(wt)
		if ctxBlock == "" {
			t.Fatal("reflectionContext is empty after 2 writes")
		}
		for _, e := range entries {
			if !strings.Contains(ctxBlock, e.failure) {
				t.Errorf("reflectionContext missing failure snippet %q", e.failure)
			}
		}
		// each attempt label must appear as a distinct heading
		for _, e := range entries {
			if !strings.Contains(ctxBlock, reflectEntryMarker+e.label) {
				t.Errorf("reflectionContext missing heading for attempt %s", e.label)
			}
		}
	})
}

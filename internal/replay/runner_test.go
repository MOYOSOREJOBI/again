package replay

import (
	"errors"
	"testing"
)

func TestRunnerLifecycle(t *testing.T) {
	Put(&Job{ID: "a", Status: "queued"})
	Run("a", func(string) error { return nil })
	if Get("a").Status != "completed" {
		t.Fatalf("expected completed")
	}
	Put(&Job{ID: "b", Status: "queued"})
	Run("b", func(string) error { return errors.New("x") })
	if Get("b").Status != "failed" {
		t.Fatalf("expected failed")
	}
}

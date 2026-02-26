package replay

import (
	"testing"
	"time"
)

func TestSortTicksOrdersByEventAndSequence(t *testing.T) {
	t1 := time.Now().UTC()
	ticks := []Tick{{EventID: "b", EventTime: t1, SequenceID: 2}, {EventID: "a", EventTime: t1, SequenceID: 1}, {EventID: "c", EventTime: t1.Add(time.Second), SequenceID: 1}}
	sortTicks(ticks)
	if ticks[0].EventID != "a" || ticks[1].EventID != "b" || ticks[2].EventID != "c" {
		t.Fatalf("unexpected order: %#v", ticks)
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

package monitor

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestTracker(t *testing.T) *Tracker {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	tracker, err := NewTracker(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tracker.Close() })
	return tracker
}

func TestRecord_And_Stats(t *testing.T) {
	tr := newTestTracker(t)

	events := []Event{
		{SkillName: "frontend-design", CLI: "cursor", Project: "proj-a", EventType: "use"},
		{SkillName: "frontend-design", CLI: "cursor", Project: "proj-a", EventType: "use"},
		{SkillName: "frontend-design", CLI: "claude", Project: "proj-b", EventType: "use"},
		{SkillName: "resume-tailoring", CLI: "cursor", Project: "proj-a", EventType: "use"},
	}

	for _, ev := range events {
		if err := tr.Record(ev); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := tr.Stats()
	if err != nil {
		t.Fatal(err)
	}

	// TopSkills
	if stats.TopSkills["frontend-design"] != 3 {
		t.Errorf("expected frontend-design count=3, got %d", stats.TopSkills["frontend-design"])
	}
	if stats.TopSkills["resume-tailoring"] != 1 {
		t.Errorf("expected resume-tailoring count=1, got %d", stats.TopSkills["resume-tailoring"])
	}

	// ByCLI
	if stats.ByCLI["cursor"] != 3 {
		t.Errorf("expected cursor count=3, got %d", stats.ByCLI["cursor"])
	}
	if stats.ByCLI["claude"] != 1 {
		t.Errorf("expected claude count=1, got %d", stats.ByCLI["claude"])
	}

	// Recent
	if len(stats.Recent) != 4 {
		t.Errorf("expected 4 recent events, got %d", len(stats.Recent))
	}
}

func TestRecord_AutoTimestamp(t *testing.T) {
	tr := newTestTracker(t)

	before := time.Now().Add(-time.Second)
	if err := tr.Record(Event{SkillName: "test", CLI: "cursor", Project: "p", EventType: "use"}); err != nil {
		t.Fatal(err)
	}
	after := time.Now().Add(time.Second)

	stats, _ := tr.Stats()
	if len(stats.Recent) != 1 {
		t.Fatal("expected 1 recent event")
	}
	ts := stats.Recent[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("auto-timestamp %v not within expected range", ts)
	}
}

func TestStats_Empty(t *testing.T) {
	tr := newTestTracker(t)

	stats, err := tr.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if len(stats.TopSkills) != 0 {
		t.Errorf("expected empty TopSkills, got %v", stats.TopSkills)
	}
	if len(stats.Recent) != 0 {
		t.Errorf("expected empty Recent, got %v", stats.Recent)
	}
	if len(stats.Stale) != 0 {
		t.Errorf("expected empty Stale, got %v", stats.Stale)
	}
}

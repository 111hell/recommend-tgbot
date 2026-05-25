package recommend

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadHistoryReturnsEmptyHistoryWhenFileMissing(t *testing.T) {
	history, err := LoadHistory(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("LoadHistory returned error: %v", err)
	}
	if len(history.Items) != 0 {
		t.Fatalf("len(history.Items) = %d, want 0", len(history.Items))
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "recommendations.json")
	want := History{Items: []HistoryItem{{
		FullName:      "owner/repo",
		RecommendedAt: "2026-05-25",
		Score:         123.45,
		Categories:    []Category{CategoryGo},
	}}}

	if err := SaveHistory(path, want); err != nil {
		t.Fatalf("SaveHistory returned error: %v", err)
	}

	got, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory returned error: %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("len(got.Items) = %d, want 1", len(got.Items))
	}
	if got.Items[0].FullName != "owner/repo" {
		t.Fatalf("FullName = %q, want owner/repo", got.Items[0].FullName)
	}
}

func TestWasRecommendedWithinUsesDayWindow(t *testing.T) {
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	history := History{Items: []HistoryItem{
		{FullName: "owner/recent", RecommendedAt: "2026-05-20"},
		{FullName: "owner/old", RecommendedAt: "2026-01-01"},
	}}

	if !history.WasRecommendedWithin("owner/recent", now, 60) {
		t.Fatal("recent item should be inside 60 day window")
	}
	if history.WasRecommendedWithin("owner/old", now, 60) {
		t.Fatal("old item should be outside 60 day window")
	}
}

func TestAddAppendsRecommendationsWithCurrentDate(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC)
	history := History{}

	history.Add([]Recommendation{{
		Repository: Repository{FullName: "owner/repo"},
		Categories: []Category{CategoryAIAgent},
		Score:      99,
	}}, now)

	if len(history.Items) != 1 {
		t.Fatalf("len(history.Items) = %d, want 1", len(history.Items))
	}
	if history.Items[0].RecommendedAt != "2026-05-25" {
		t.Fatalf("RecommendedAt = %q, want 2026-05-25", history.Items[0].RecommendedAt)
	}
}

func TestLoadHistoryRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recommendations.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write history: %v", err)
	}

	if _, err := LoadHistory(path); err == nil {
		t.Fatal("LoadHistory returned nil error, want invalid JSON error")
	}
}

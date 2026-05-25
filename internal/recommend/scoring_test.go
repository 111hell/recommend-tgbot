package recommend

import (
	"testing"
	"time"
)

func TestScoreRepositoryPrefersGoWhenQualityIsSimilar(t *testing.T) {
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	goRepo := Repository{
		FullName:    "owner/go-tool",
		Language:    "Go",
		Description: "Useful Go infrastructure toolkit.",
		Topics:      []string{"golang"},
		Stars:       1000,
		Forks:       100,
		Watchers:    1000,
		LicenseName: "MIT",
		PushedAt:    now.AddDate(0, 0, -3),
	}
	otherRepo := goRepo
	otherRepo.FullName = "owner/python-tool"
	otherRepo.Language = "Python"
	otherRepo.Description = "Useful infrastructure toolkit."
	otherRepo.Topics = []string{"tooling"}

	if ScoreRepository(goRepo, History{}, now) <= ScoreRepository(otherRepo, History{}, now) {
		t.Fatal("Go repository should score higher than a similar general repository")
	}
}

func TestScoreRepositoryRewardsRecentActivity(t *testing.T) {
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	active := Repository{
		FullName:    "owner/active",
		Language:    "Go",
		Description: "Active Go project.",
		Stars:       1000,
		Forks:       100,
		LicenseName: "Apache-2.0",
		PushedAt:    now.AddDate(0, 0, -2),
	}
	stale := active
	stale.FullName = "owner/stale"
	stale.PushedAt = now.AddDate(-2, 0, 0)

	if ScoreRepository(active, History{}, now) <= ScoreRepository(stale, History{}, now) {
		t.Fatal("recently pushed repository should score higher than stale repository")
	}
}

func TestScoreRepositoryPenalizesRecentRecommendation(t *testing.T) {
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	repo := Repository{
		FullName:    "owner/repeated",
		Language:    "Go",
		Description: "Repeated project.",
		Stars:       2000,
		Forks:       200,
		LicenseName: "MIT",
		PushedAt:    now.AddDate(0, 0, -1),
	}
	history := History{Items: []HistoryItem{{
		FullName:      "owner/repeated",
		RecommendedAt: "2026-05-20",
		Score:         500,
		Categories:    []Category{CategoryGo},
	}}}

	if ScoreRepository(repo, history, now) >= ScoreRepository(repo, History{}, now) {
		t.Fatal("recently recommended repository should be penalized")
	}
}

func TestRankRecommendationsFiltersArchivedForksAndExcludedRepos(t *testing.T) {
	now := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	repos := []Repository{
		{FullName: "owner/keep", Language: "Go", Stars: 1000, PushedAt: now},
		{FullName: "owner/archive", Language: "Go", Stars: 5000, Archived: true, PushedAt: now},
		{FullName: "owner/fork", Language: "Go", Stars: 5000, Fork: true, PushedAt: now},
		{FullName: "owner/exclude", Language: "Go", Stars: 5000, PushedAt: now},
	}

	recommendations := RankRecommendations(repos, History{}, RankingOptions{
		Count:        5,
		MinStars:     500,
		ExcludeRepos: []string{"owner/exclude"},
		Now:          now,
	})

	if len(recommendations) != 1 {
		t.Fatalf("len(recommendations) = %d, want 1", len(recommendations))
	}
	if recommendations[0].Repository.FullName != "owner/keep" {
		t.Fatalf("recommended %q, want owner/keep", recommendations[0].Repository.FullName)
	}
}

package recommend

import (
	"math"
	"sort"
	"strings"
	"time"
)

type RankingOptions struct {
	Count        int
	MinStars     int
	ExcludeRepos []string
	Now          time.Time
}

func ScoreRepository(repo Repository, history History, now time.Time) float64 {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	score := math.Log10(float64(max(repo.Stars, 1))) * 100
	score += math.Log10(float64(max(repo.Forks, 1))) * 20
	score += math.Log10(float64(max(repo.Watchers, repo.Stars, 1))) * 10
	score += repo.SearchRankBoost

	daysSincePush := now.Sub(repo.PushedAt).Hours() / 24
	switch {
	case repo.PushedAt.IsZero():
		score -= 30
	case daysSincePush <= 7:
		score += 60
	case daysSincePush <= 30:
		score += 40
	case daysSincePush <= 180:
		score += 15
	case daysSincePush > 365:
		score -= 60
	}

	categories := DetectCategories(repo)
	for _, category := range categories {
		switch category {
		case CategoryGo:
			score += 45
		case CategoryAIAgent:
			score += 35
		case CategoryDevOpsInfra:
			score += 30
		case CategoryGeneral:
			score -= 25
		}
	}

	if strings.TrimSpace(repo.Description) != "" {
		score += 15
	}
	if strings.TrimSpace(repo.LicenseName) != "" {
		score += 15
	}
	if strings.TrimSpace(repo.Homepage) != "" {
		score += 8
	}
	if len(repo.Topics) >= 3 {
		score += 10
	}
	if repo.OpenIssues > 0 && repo.Stars > 0 {
		ratio := float64(repo.OpenIssues) / float64(repo.Stars)
		if ratio > 0.2 {
			score -= 20
		}
	}
	if history.WasRecommendedWithin(repo.FullName, now, 60) {
		score -= 500
	}

	return score
}

func RankRecommendations(repos []Repository, history History, opts RankingOptions) []Recommendation {
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	if opts.Count <= 0 {
		opts.Count = 5
	}

	excluded := make(map[string]struct{}, len(opts.ExcludeRepos))
	for _, repo := range opts.ExcludeRepos {
		excluded[strings.ToLower(strings.TrimSpace(repo))] = struct{}{}
	}

	recommendations := make([]Recommendation, 0, len(repos))
	seen := map[string]struct{}{}
	for _, repo := range repos {
		key := strings.ToLower(repo.FullName)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, ok := excluded[key]; ok {
			continue
		}
		if repo.Archived || repo.Fork || repo.Stars < opts.MinStars {
			continue
		}
		score := ScoreRepository(repo, history, opts.Now)
		recommendations = append(recommendations, Recommendation{
			Repository: repo,
			Categories: DetectCategories(repo),
			Score:      score,
			Reason:     BuildReason(repo, DetectCategories(repo)),
		})
	}

	sort.SliceStable(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})
	if len(recommendations) > opts.Count {
		recommendations = recommendations[:opts.Count]
	}
	return recommendations
}

func BuildReason(repo Repository, categories []Category) string {
	if categoryExists(categories, CategoryGo) {
		return "Go 相关项目，适合学习工程化实现和后端/基础设施实践。"
	}
	if categoryExists(categories, CategoryAIAgent) {
		return "AI/Agent 方向项目，适合跟进 LLM 应用、工作流和工具链实践。"
	}
	if categoryExists(categories, CategoryDevOpsInfra) {
		return "运维/基础设施方向项目，适合学习部署、可观测性或自动化实践。"
	}
	return "高质量工程项目，适合学习成熟开源项目的设计和实现。"
}

func max(values ...int) int {
	result := values[0]
	for _, value := range values[1:] {
		if value > result {
			result = value
		}
	}
	return result
}

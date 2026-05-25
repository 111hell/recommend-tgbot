package recommend

import (
	"fmt"
	"strings"
)

func FormatMessage(recommendations []Recommendation) string {
	if len(recommendations) == 0 {
		return "今天没有找到符合条件的 GitHub 项目。"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "今日推荐：%d 个值得学习的 GitHub 项目\n", len(recommendations))

	for i, recommendation := range recommendations {
		repo := recommendation.Repository
		fmt.Fprintf(&b, "\n%d. %s\n", i+1, repo.FullName)
		fmt.Fprintf(&b, "分类：%s\n", FormatCategories(recommendation.Categories))
		if strings.TrimSpace(repo.Description) != "" {
			fmt.Fprintf(&b, "简介：%s\n", strings.TrimSpace(repo.Description))
		}
		fmt.Fprintf(&b, "指标：Stars：%d / Forks：%d", repo.Stars, repo.Forks)
		if strings.TrimSpace(repo.Language) != "" {
			fmt.Fprintf(&b, " / Language：%s", repo.Language)
		}
		b.WriteString("\n")
		if strings.TrimSpace(recommendation.Reason) != "" {
			fmt.Fprintf(&b, "推荐理由：%s\n", strings.TrimSpace(recommendation.Reason))
		}
		fmt.Fprintf(&b, "链接：%s\n", repoURL(repo))
	}

	return b.String()
}

func FormatCategories(categories []Category) string {
	if len(categories) == 0 {
		return "General"
	}
	labels := make([]string, 0, len(categories))
	for _, category := range categories {
		switch category {
		case CategoryGo:
			labels = append(labels, "Go")
		case CategoryAIAgent:
			labels = append(labels, "AI Agent")
		case CategoryDevOpsInfra:
			labels = append(labels, "DevOps Infra")
		case CategoryGeneral:
			labels = append(labels, "General")
		default:
			labels = append(labels, string(category))
		}
	}
	return strings.Join(labels, " / ")
}

func repoURL(repo Repository) string {
	if strings.TrimSpace(repo.HTMLURL) != "" {
		return strings.TrimSpace(repo.HTMLURL)
	}
	return "https://github.com/" + repo.FullName
}

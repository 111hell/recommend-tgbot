package recommend

import (
	"strings"
	"testing"
)

func TestFormatMessageIncludesChineseSummaryAndRepoDetails(t *testing.T) {
	message := FormatMessage([]Recommendation{{
		Repository: Repository{
			FullName:    "owner/repo",
			Description: "Useful Go agent toolkit.",
			HTMLURL:     "https://github.com/owner/repo",
			Stars:       1200,
			Forks:       80,
			Language:    "Go",
		},
		Categories: []Category{CategoryGo, CategoryAIAgent},
		Score:      321.5,
		Reason:     "Go 相关项目，适合学习工程化实现和后端/基础设施实践。",
	}})

	for _, want := range []string{
		"今日推荐：1 个值得学习的 GitHub 项目",
		"owner/repo",
		"分类：Go / AI Agent",
		"Stars：1200",
		"Forks：80",
		"推荐理由：Go 相关项目",
		"链接：https://github.com/owner/repo",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("message missing %q:\n%s", want, message)
		}
	}
}

func TestFormatMessageHandlesEmptyRecommendations(t *testing.T) {
	message := FormatMessage(nil)
	if !strings.Contains(message, "今天没有找到符合条件的 GitHub 项目") {
		t.Fatalf("unexpected empty message: %s", message)
	}
}

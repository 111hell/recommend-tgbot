package learning

import (
	"strings"
	"unicode"

	"recommend-bot/internal/recommend"
)

type Plan struct {
	FullName        string
	Slug            string
	Description     string
	GitHubURL       string
	Language        string
	Categories      []recommend.Category
	Markdown        string
	WhyLearn        string
	Highlights      []string
	HardParts       []string
	Days            []DayPlan
	ReviewQuestions []string
}

type DayPlan struct {
	Day       int
	Title     string
	Checklist []string
}

func BuildTemplatePlan(rec recommend.Recommendation) Plan {
	repo := rec.Repository
	why := strings.TrimSpace(rec.Reason)
	if why == "" {
		why = recommend.BuildReason(repo, rec.Categories)
	}

	return Plan{
		FullName:    repo.FullName,
		Slug:        RepoSlug(repo.FullName),
		Description: strings.TrimSpace(repo.Description),
		GitHubURL:   repoURL(repo),
		Language:    strings.TrimSpace(repo.Language),
		Categories:  rec.Categories,
		WhyLearn:    why,
		Highlights: []string{
			"梳理项目解决的问题、目标用户和核心使用场景。",
			"观察项目如何拆分模块、暴露 API、组织示例和文档。",
			"提炼一个可以迁移到自己项目里的工程实践。",
		},
		HardParts: []string{
			"判断核心抽象为什么这样设计，而不是只阅读表层用法。",
			"区分框架能力、业务约束和历史演进留下的复杂度。",
			"从 README、示例和源码入口之间建立同一张心智地图。",
		},
		Days: []DayPlan{
			{
				Day:   1,
				Title: "建立项目全局认知",
				Checklist: []string{
					"阅读 README，写下项目解决的问题和最典型的使用场景。",
					"浏览顶层目录，标记源码、文档、示例和测试分别在哪里。",
					"运行或阅读一个最小示例，确认项目的主要入口和基本调用方式。",
					"记录 3 个你认为值得深入的问题。",
				},
			},
			{
				Day:   2,
				Title: "拆解核心设计和可迁移经验",
				Checklist: []string{
					"选择一个核心模块，阅读它的主要类型、接口和调用链。",
					"总结这个模块的设计取舍：它优化了什么，又牺牲了什么。",
					"找一处错误处理、扩展点、测试或性能相关实现，说明它解决的问题。",
					"写一个小结：这个项目有什么做法可以用于自己的开发工作。",
				},
			},
		},
		ReviewQuestions: []string{
			"这个项目的核心抽象是什么？",
			"它的目录结构如何帮助维护者控制复杂度？",
			"如果让你实现一个最小版本，你会保留哪些设计，删掉哪些设计？",
			"这个项目里最值得迁移到自己项目的工程实践是什么？",
		},
	}
}

func RepoSlug(fullName string) string {
	parts := strings.Split(strings.TrimSpace(fullName), "/")
	name := parts[len(parts)-1]
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func repoURL(repo recommend.Repository) string {
	if strings.TrimSpace(repo.HTMLURL) != "" {
		return strings.TrimSpace(repo.HTMLURL)
	}
	return "https://github.com/" + repo.FullName
}

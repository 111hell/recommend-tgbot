package learning

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"recommend-bot/internal/recommend"
)

type CodexOptions struct {
	Command string
	Timeout time.Duration
}

func BuildCodexPlan(ctx context.Context, rec recommend.Recommendation, opts CodexOptions) (Plan, error) {
	base := BuildTemplatePlan(rec)
	command := strings.TrimSpace(opts.Command)
	if command == "" {
		command = "codex"
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	outputFile, err := os.CreateTemp("", "recommend-bot-codex-*.md")
	if err != nil {
		return base, fmt.Errorf("create codex output file: %w", err)
	}
	outputPath := outputFile.Name()
	_ = outputFile.Close()
	defer os.Remove(outputPath)

	cmd := exec.CommandContext(runCtx, command, "exec", "--ephemeral", "--sandbox", "read-only", "-o", outputPath, codexPrompt(base, rec))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return base, fmt.Errorf("codex analysis failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	finalMessage, err := os.ReadFile(outputPath)
	if err != nil {
		return base, fmt.Errorf("read codex output: %w", err)
	}

	markdown := strings.TrimSpace(string(finalMessage))
	if markdown == "" {
		return base, fmt.Errorf("codex analysis returned empty output")
	}
	base.Markdown = ensureTitle(markdown, base.FullName)
	return base, nil
}

func codexPrompt(plan Plan, rec recommend.Recommendation) string {
	return fmt.Sprintf(`你是一个资深开源项目学习教练。请基于下面的 GitHub 项目信息，生成一份适合写入 Obsidian 的中文 Markdown 学习报告。

目标：
- 不要泛泛而谈，不要只说“Go 项目适合学习工程化”。
- 具体解释这个项目为什么值得学，框架/架构为什么这样设计，好处和代价是什么。
- 把所有最值得学习的亮点压缩进两天内完成。
- 输出完整 Markdown，不要输出代码块围栏。

项目：
- full_name: %s
- url: %s
- language: %s
- description: %s
- categories: %s
- stars: %d
- forks: %d
- baseline_reason: %s

必须包含这些标题：
# %s
## 为什么值得学
## 项目解决什么问题
## 架构和设计亮点
## 业务/工程重难点
## Day 1：建立项目全局认知
## Day 2：拆解核心设计和可迁移经验
## 复盘问题
## 我的笔记

Day 1 和 Day 2 下面必须用 Markdown checkbox：
- [ ] ...

“我的笔记”下面只放一个空 checkbox 或短横线，留给用户自己写。`,
		plan.FullName,
		plan.GitHubURL,
		plan.Language,
		plan.Description,
		recommend.FormatCategories(plan.Categories),
		rec.Repository.Stars,
		rec.Repository.Forks,
		rec.Reason,
		plan.FullName,
	)
}

func ensureTitle(markdown string, fullName string) string {
	if strings.HasPrefix(strings.TrimSpace(markdown), "# ") {
		return strings.TrimSpace(markdown) + "\n"
	}
	return "# " + strings.TrimSpace(fullName) + "\n\n" + strings.TrimSpace(markdown) + "\n"
}

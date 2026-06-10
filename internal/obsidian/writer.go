package obsidian

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"recommend-bot/internal/learning"
	"recommend-bot/internal/recommend"
)

type Writer struct {
	VaultPath  string
	ProjectDir string
}

type WriteResult struct {
	RelativePath  string
	AbsolutePath  string
	DeepLink      string
	AlreadyExists bool
}

func (w Writer) WritePlan(plan learning.Plan) (WriteResult, error) {
	if strings.TrimSpace(w.VaultPath) == "" {
		return WriteResult{}, fmt.Errorf("obsidian vault path is required")
	}
	if info, err := os.Stat(w.VaultPath); err != nil {
		return WriteResult{}, fmt.Errorf("obsidian vault path is not accessible: %w", err)
	} else if !info.IsDir() {
		return WriteResult{}, fmt.Errorf("obsidian vault path is not a directory: %s", w.VaultPath)
	}

	projectDir := strings.TrimSpace(w.ProjectDir)
	if projectDir == "" {
		projectDir = "GitHub Projects"
	}
	slug := strings.TrimSpace(plan.Slug)
	if slug == "" {
		slug = learning.RepoSlug(plan.FullName)
	}
	relativePath := filepath.Join(projectDir, slug, slug+".md")
	absolutePath := filepath.Join(w.VaultPath, relativePath)
	result := WriteResult{
		RelativePath: relativePath,
		AbsolutePath: absolutePath,
		DeepLink:     DeepLink(filepath.Base(w.VaultPath), relativePath),
	}

	if _, err := os.Stat(absolutePath); err == nil {
		result.AlreadyExists = true
		return result, nil
	} else if !os.IsNotExist(err) {
		return WriteResult{}, fmt.Errorf("check obsidian note: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("create obsidian project dir: %w", err)
	}
	if err := os.WriteFile(absolutePath, []byte(RenderMarkdown(plan)), 0o644); err != nil {
		return WriteResult{}, fmt.Errorf("write obsidian note: %w", err)
	}
	return result, nil
}

func RenderMarkdown(plan learning.Plan) string {
	var b strings.Builder
	title := strings.TrimSpace(plan.FullName)
	if title == "" {
		title = strings.TrimSpace(plan.Slug)
	}
	fmt.Fprintf(&b, "# %s\n\n", title)

	if strings.TrimSpace(plan.GitHubURL) != "" {
		fmt.Fprintf(&b, "- GitHub: [%s](%s)\n", title, strings.TrimSpace(plan.GitHubURL))
	}
	if strings.TrimSpace(plan.Language) != "" {
		fmt.Fprintf(&b, "- Language: %s\n", strings.TrimSpace(plan.Language))
	}
	if len(plan.Categories) > 0 {
		fmt.Fprintf(&b, "- Categories: %s\n", recommend.FormatCategories(plan.Categories))
	}
	if strings.TrimSpace(plan.Description) != "" {
		fmt.Fprintf(&b, "- Description: %s\n", strings.TrimSpace(plan.Description))
	}

	writeSection(&b, "为什么值得学", []string{plan.WhyLearn})
	writeSection(&b, "项目亮点", plan.Highlights)
	writeSection(&b, "重难点", plan.HardParts)

	for _, day := range plan.Days {
		fmt.Fprintf(&b, "\n## Day %d：%s\n\n", day.Day, strings.TrimSpace(day.Title))
		for _, item := range day.Checklist {
			if strings.TrimSpace(item) != "" {
				fmt.Fprintf(&b, "- [ ] %s\n", strings.TrimSpace(item))
			}
		}
	}

	writeSection(&b, "复盘问题", plan.ReviewQuestions)
	b.WriteString("\n## 我的笔记\n\n")
	b.WriteString("- \n")
	return b.String()
}

func DeepLink(vaultName string, relativePath string) string {
	return "obsidian://open?vault=" + url.QueryEscape(vaultName) + "&file=" + url.QueryEscape(filepath.ToSlash(relativePath))
}

func writeSection(b *strings.Builder, title string, items []string) {
	var cleaned []string
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			cleaned = append(cleaned, strings.TrimSpace(item))
		}
	}
	if len(cleaned) == 0 {
		return
	}
	fmt.Fprintf(b, "\n## %s\n\n", title)
	for _, item := range cleaned {
		fmt.Fprintf(b, "- %s\n", item)
	}
}

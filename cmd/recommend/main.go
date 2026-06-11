package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"recommend-bot/internal/config"
	gh "recommend-bot/internal/github"
	"recommend-bot/internal/learning"
	"recommend-bot/internal/obsidian"
	"recommend-bot/internal/recommend"
	"recommend-bot/internal/telegram"
)

const historyPath = "data/recommendations.json"
const learningStatePath = "data/learning_state.json"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "recommend-bot: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.yaml", "path to config yaml")
	dryRun := flag.Bool("dry-run", false, "print message without sending to Telegram or saving history")
	obsidianMode := flag.Bool("obsidian", false, "write a two-day learning route to Obsidian")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	env, err := config.LoadEnv(*dryRun)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	history, err := recommend.LoadHistory(historyPath)
	if err != nil {
		return err
	}

	var candidates []recommend.Repository
	if *dryRun && env.GitHubToken == "" {
		candidates = sampleCandidates()
	} else {
		githubClient := gh.NewClient(env.GitHubToken)
		candidates, err = githubClient.SearchCandidates(ctx, cfg.Recommend.MinStars, cfg.Recommend.LookbackDays)
		if err != nil {
			return err
		}
	}

	recommendations := recommend.RankRecommendations(candidates, history, recommend.RankingOptions{
		Count:        cfg.Recommend.Count,
		MinStars:     cfg.Recommend.MinStars,
		ExcludeRepos: cfg.Recommend.ExcludeRepos,
		Now:          time.Now().UTC(),
	})

	if *obsidianMode {
		return runObsidianMode(cfg, env, recommendations, history, *dryRun)
	}

	message := recommend.FormatMessage(recommendations)
	if *dryRun {
		fmt.Print(message)
		return nil
	}

	telegramClient := telegram.NewClient(env.TelegramBotToken)
	if err := telegramClient.SendMessage(ctx, env.TelegramChatID, message, cfg.Telegram.ParseMode); err != nil {
		return err
	}

	history.Add(recommendations, time.Now().UTC())
	if err := recommend.SaveHistory(historyPath, history); err != nil {
		return err
	}

	fmt.Printf("sent %d recommendations\n", len(recommendations))
	return nil
}

func runObsidianMode(cfg config.Config, env config.Env, recommendations []recommend.Recommendation, history recommend.History, dryRun bool) error {
	now := nowInShanghai()
	state, err := learning.LoadState(learningStatePath)
	if err != nil {
		return err
	}

	if !state.NeedsNewProject(now) {
		route := state.Active
		day := state.CurrentRouteDay(now)
		plan := learning.Plan{
			FullName:  route.FullName,
			Slug:      route.Slug,
			GitHubURL: "https://github.com/" + route.FullName,
		}
		result := obsidian.WriteResult{
			RelativePath: route.ObsidianPath,
			DeepLink:     obsidian.DeepLinkForPath(joinPath(cfg.Obsidian.VaultPath, route.ObsidianPath)),
		}
		message := buildLearningReminder(plan, result, day)
		if dryRun {
			fmt.Print(message)
			return nil
		}
		telegramClient := telegram.NewClient(env.TelegramBotToken)
		return telegramClient.SendMessageWithOptions(context.Background(), env.TelegramChatID, message, telegram.SendOptions{
			ParseMode: cfg.Telegram.ParseMode,
			Buttons:   reminderButtons(plan, result),
		})
	}

	if len(recommendations) == 0 {
		message := "今天没有找到符合条件的 GitHub 学习项目。"
		if dryRun {
			fmt.Print(message)
			return nil
		}
		telegramClient := telegram.NewClient(env.TelegramBotToken)
		return telegramClient.SendMessage(context.Background(), env.TelegramChatID, message, cfg.Telegram.ParseMode)
	}

	plan, analysisErr := buildPlan(context.Background(), cfg, recommendations[0])
	writer := obsidian.Writer{VaultPath: cfg.Obsidian.VaultPath, ProjectDir: cfg.Obsidian.ProjectDir}
	var result obsidian.WriteResult
	if dryRun {
		relativePath := cfg.Obsidian.ProjectDir + "/" + plan.Slug + "/" + plan.Slug + ".md"
		result = obsidian.WriteResult{
			RelativePath: relativePath,
			DeepLink:     obsidian.DeepLinkForPath(joinPath(cfg.Obsidian.VaultPath, relativePath)),
		}
		fmt.Print(obsidian.RenderMarkdown(plan))
		if analysisErr != nil {
			fmt.Fprintf(os.Stderr, "analysis fallback: %v\n", analysisErr)
		}
		fmt.Print("\n--- Telegram Reminder ---\n")
		fmt.Print(buildLearningReminder(plan, result, 1))
		return nil
	}

	result, err = writer.WritePlan(plan)
	if err != nil {
		return err
	}
	state.Active = &learning.Route{
		FullName:     plan.FullName,
		Slug:         plan.Slug,
		ObsidianPath: result.RelativePath,
		GeneratedAt:  now.Format("2006-01-02"),
		CurrentDay:   1,
		Day1Due:      now.Format("2006-01-02"),
		Day2Due:      now.AddDate(0, 0, 1).Format("2006-01-02"),
		Status:       learning.StatusActive,
	}
	if err := learning.SaveState(learningStatePath, state); err != nil {
		return err
	}
	history.Add(recommendations[:1], now)
	if err := recommend.SaveHistory(historyPath, history); err != nil {
		return err
	}

	message := buildLearningReminder(plan, result, 1)
	telegramClient := telegram.NewClient(env.TelegramBotToken)
	return telegramClient.SendMessageWithOptions(context.Background(), env.TelegramChatID, message, telegram.SendOptions{
		ParseMode: cfg.Telegram.ParseMode,
		Buttons:   reminderButtons(plan, result),
	})
}

func buildPlan(ctx context.Context, cfg config.Config, rec recommend.Recommendation) (learning.Plan, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Analysis.Provider)) {
	case "codex":
		plan, err := learning.BuildCodexPlan(ctx, rec, learning.CodexOptions{})
		if err != nil {
			return learning.BuildTemplatePlan(rec), err
		}
		return plan, nil
	default:
		return learning.BuildTemplatePlan(rec), nil
	}
}

func nowInShanghai() time.Time {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Now().UTC()
	}
	return time.Now().In(location)
}

func buildLearningReminder(plan learning.Plan, result obsidian.WriteResult, day int) string {
	if day <= 0 {
		day = 1
	}
	return fmt.Sprintf("今日学习项目：%s\n\n路线已写入 Obsidian：\n%s\n\n打开 Obsidian：\n%s\n\nGitHub：\n%s\n\n今天完成 Day %d。\n",
		plan.FullName,
		result.RelativePath,
		result.DeepLink,
		plan.GitHubURL,
		day,
	)
}

func reminderButtons(plan learning.Plan, result obsidian.WriteResult) []telegram.Button {
	buttons := []telegram.Button{}
	if plan.GitHubURL != "" {
		buttons = append(buttons, telegram.Button{Text: "打开 GitHub", URL: plan.GitHubURL})
	}
	return buttons
}

func joinPath(base string, relative string) string {
	for strings.HasSuffix(base, "/") {
		base = strings.TrimSuffix(base, "/")
	}
	for strings.HasPrefix(relative, "/") {
		relative = strings.TrimPrefix(relative, "/")
	}
	if base == "" {
		return relative
	}
	if relative == "" {
		return base
	}
	return base + "/" + relative
}

func sampleCandidates() []recommend.Repository {
	now := time.Now().UTC()
	return []recommend.Repository{
		{
			FullName:    "gin-gonic/gin",
			Name:        "gin",
			Description: "Gin is a HTTP web framework written in Go.",
			HTMLURL:     "https://github.com/gin-gonic/gin",
			Language:    "Go",
			Topics:      []string{"go", "golang", "framework"},
			Stars:       80000,
			Forks:       8000,
			Watchers:    80000,
			LicenseName: "MIT",
			PushedAt:    now.AddDate(0, 0, -3),
		},
		{
			FullName:    "modelcontextprotocol/servers",
			Name:        "servers",
			Description: "Model Context Protocol servers for AI agent integrations.",
			HTMLURL:     "https://github.com/modelcontextprotocol/servers",
			Language:    "TypeScript",
			Topics:      []string{"mcp", "llm", "agents"},
			Stars:       50000,
			Forks:       5000,
			Watchers:    50000,
			LicenseName: "MIT",
			PushedAt:    now.AddDate(0, 0, -2),
		},
		{
			FullName:    "prometheus/prometheus",
			Name:        "prometheus",
			Description: "Monitoring system and time series database.",
			HTMLURL:     "https://github.com/prometheus/prometheus",
			Language:    "Go",
			Topics:      []string{"monitoring", "observability", "prometheus"},
			Stars:       60000,
			Forks:       9000,
			Watchers:    60000,
			LicenseName: "Apache-2.0",
			PushedAt:    now.AddDate(0, 0, -1),
		},
	}
}

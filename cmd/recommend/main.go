package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"recommend-bot/internal/config"
	gh "recommend-bot/internal/github"
	"recommend-bot/internal/recommend"
	"recommend-bot/internal/telegram"
)

const historyPath = "data/recommendations.json"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "recommend-bot: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.yaml", "path to config yaml")
	dryRun := flag.Bool("dry-run", false, "print message without sending to Telegram or saving history")
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

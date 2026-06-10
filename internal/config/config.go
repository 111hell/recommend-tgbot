package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Recommend RecommendConfig `yaml:"recommend"`
	Telegram  TelegramConfig  `yaml:"telegram"`
	Obsidian  ObsidianConfig  `yaml:"obsidian"`
	Analysis  AnalysisConfig  `yaml:"analysis"`
	LLM       LLMConfig       `yaml:"llm"`
}

type RecommendConfig struct {
	Count        int      `yaml:"count"`
	MinStars     int      `yaml:"min_stars"`
	LookbackDays int      `yaml:"lookback_days"`
	ExcludeRepos []string `yaml:"exclude_repos"`
	Categories   []string `yaml:"categories"`
}

type TelegramConfig struct {
	ParseMode string `yaml:"parse_mode"`
}

type ObsidianConfig struct {
	VaultPath  string `yaml:"vault_path"`
	ProjectDir string `yaml:"project_dir"`
}

type AnalysisConfig struct {
	Provider     string `yaml:"provider"`
	ScheduleTime string `yaml:"schedule_time"`
}

type LLMConfig struct {
	Provider  string `yaml:"provider"`
	BaseURL   string `yaml:"base_url"`
	Model     string `yaml:"model"`
	APIKeyEnv string `yaml:"api_key_env"`
}

type Env struct {
	TelegramBotToken string
	TelegramChatID   string
	GitHubToken      string
}

func Load(path string) (Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&cfg)

	return cfg, nil
}

func LoadEnv(dryRun bool) (Env, error) {
	env := Env{
		TelegramBotToken: strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		TelegramChatID:   strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")),
		GitHubToken:      strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
	}

	var missing []string
	if !dryRun && env.GitHubToken == "" {
		missing = append(missing, "GITHUB_TOKEN")
	}
	if !dryRun {
		if env.TelegramBotToken == "" {
			missing = append(missing, "TELEGRAM_BOT_TOKEN")
		}
		if env.TelegramChatID == "" {
			missing = append(missing, "TELEGRAM_CHAT_ID")
		}
	}
	if len(missing) > 0 {
		return Env{}, errors.New("missing required environment variables: " + strings.Join(missing, ", "))
	}

	return env, nil
}

func defaultConfig() Config {
	return Config{
		Recommend: RecommendConfig{
			Count:        5,
			MinStars:     500,
			LookbackDays: 30,
			Categories:   []string{"go", "ai_agent", "devops_infra"},
		},
		Telegram: TelegramConfig{
			ParseMode: "Markdown",
		},
		Obsidian: ObsidianConfig{
			VaultPath:  "/Users/janeshirley/Documents/Obsidian Vault",
			ProjectDir: "GitHub Projects",
		},
		Analysis: AnalysisConfig{
			Provider:     "template",
			ScheduleTime: "11:00",
		},
		LLM: LLMConfig{
			Provider:  "openai_compatible",
			APIKeyEnv: "LLM_API_KEY",
		},
	}
}

func applyDefaults(cfg *Config) {
	defaults := defaultConfig()
	if cfg.Recommend.Count == 0 {
		cfg.Recommend.Count = defaults.Recommend.Count
	}
	if cfg.Recommend.MinStars == 0 {
		cfg.Recommend.MinStars = defaults.Recommend.MinStars
	}
	if cfg.Recommend.LookbackDays == 0 {
		cfg.Recommend.LookbackDays = defaults.Recommend.LookbackDays
	}
	if len(cfg.Recommend.Categories) == 0 {
		cfg.Recommend.Categories = defaults.Recommend.Categories
	}
	if strings.TrimSpace(cfg.Telegram.ParseMode) == "" {
		cfg.Telegram.ParseMode = defaults.Telegram.ParseMode
	}
	if strings.TrimSpace(cfg.Obsidian.VaultPath) == "" {
		cfg.Obsidian.VaultPath = defaults.Obsidian.VaultPath
	}
	if strings.TrimSpace(cfg.Obsidian.ProjectDir) == "" {
		cfg.Obsidian.ProjectDir = defaults.Obsidian.ProjectDir
	}
	if strings.TrimSpace(cfg.Analysis.Provider) == "" {
		cfg.Analysis.Provider = defaults.Analysis.Provider
	}
	if strings.TrimSpace(cfg.Analysis.ScheduleTime) == "" {
		cfg.Analysis.ScheduleTime = defaults.Analysis.ScheduleTime
	}
	if strings.TrimSpace(cfg.LLM.Provider) == "" {
		cfg.LLM.Provider = defaults.LLM.Provider
	}
	if strings.TrimSpace(cfg.LLM.APIKeyEnv) == "" {
		cfg.LLM.APIKeyEnv = defaults.LLM.APIKeyEnv
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsYAMLConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
recommend:
  count: 3
  min_stars: 750
  lookback_days: 45
  exclude_repos:
    - owner/skip
  categories:
    - go
    - ai_agent
telegram:
  parse_mode: Markdown
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Recommend.Count != 3 {
		t.Fatalf("Count = %d, want 3", cfg.Recommend.Count)
	}
	if cfg.Recommend.MinStars != 750 {
		t.Fatalf("MinStars = %d, want 750", cfg.Recommend.MinStars)
	}
	if cfg.Recommend.LookbackDays != 45 {
		t.Fatalf("LookbackDays = %d, want 45", cfg.Recommend.LookbackDays)
	}
	if got := cfg.Recommend.ExcludeRepos[0]; got != "owner/skip" {
		t.Fatalf("ExcludeRepos[0] = %q, want owner/skip", got)
	}
	if got := cfg.Recommend.Categories[1]; got != "ai_agent" {
		t.Fatalf("Categories[1] = %q, want ai_agent", got)
	}
	if cfg.Telegram.ParseMode != "Markdown" {
		t.Fatalf("ParseMode = %q, want Markdown", cfg.Telegram.ParseMode)
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Recommend.Count != 5 {
		t.Fatalf("Count = %d, want 5", cfg.Recommend.Count)
	}
	if cfg.Recommend.MinStars != 500 {
		t.Fatalf("MinStars = %d, want 500", cfg.Recommend.MinStars)
	}
	if cfg.Recommend.LookbackDays != 30 {
		t.Fatalf("LookbackDays = %d, want 30", cfg.Recommend.LookbackDays)
	}
	if cfg.Telegram.ParseMode != "Markdown" {
		t.Fatalf("ParseMode = %q, want Markdown", cfg.Telegram.ParseMode)
	}
}

func TestLoadEnvRequiresSecretsWhenNotDryRun(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("GITHUB_TOKEN", "")

	_, err := LoadEnv(false)
	if err == nil {
		t.Fatal("LoadEnv returned nil error, want missing secret error")
	}
}

func TestLoadEnvAllowsMissingSecretsForDryRun(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	t.Setenv("GITHUB_TOKEN", "")

	env, err := LoadEnv(true)
	if err != nil {
		t.Fatalf("LoadEnv returned error: %v", err)
	}
	if env.GitHubToken != "" {
		t.Fatalf("GitHubToken = %q, want empty", env.GitHubToken)
	}
}

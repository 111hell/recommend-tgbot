# GitHub Recommendation Bot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI that recommends high-quality GitHub projects daily and sends the result to a Telegram private chat through GitHub Actions.

**Architecture:** The CLI loads public settings from `config.yaml` and secrets from environment variables, queries GitHub for category-specific candidate repositories, scores and de-duplicates them, formats a Chinese recommendation message, sends it via Telegram, and writes public recommendation history to `data/recommendations.json`. GitHub Actions runs the CLI daily at 05:00 Asia/Shanghai and commits history updates after successful sends.

**Tech Stack:** Go 1.22, standard `net/http`, `encoding/json`, `gopkg.in/yaml.v3`, GitHub REST Search API, Telegram Bot API, GitHub Actions.

---

## File Structure

- Create `go.mod`: module definition and YAML dependency.
- Create `config.yaml`: public recommendation defaults.
- Create `cmd/recommend/main.go`: CLI entrypoint and orchestration.
- Create `internal/config/config.go`: YAML loading and environment validation.
- Create `internal/config/config_test.go`: config behavior tests.
- Create `internal/recommend/model.go`: shared repository, category, and recommendation types.
- Create `internal/recommend/category.go`: category detection from GitHub metadata.
- Create `internal/recommend/category_test.go`: category detection tests.
- Create `internal/recommend/scoring.go`: deterministic scoring and candidate ranking.
- Create `internal/recommend/scoring_test.go`: scoring tests.
- Create `internal/recommend/history.go`: history load/save and recent duplicate checks.
- Create `internal/recommend/history_test.go`: history tests.
- Create `internal/recommend/message.go`: Chinese Telegram message formatting.
- Create `internal/recommend/message_test.go`: message formatting tests.
- Create `internal/github/client.go`: GitHub Search API client and query builder.
- Create `internal/telegram/client.go`: Telegram Bot API sender.
- Create `.github/workflows/daily.yml`: scheduled and manual workflow.
- Create `data/.gitkeep`: keep the history directory in git before the first run.
- Create `README.md`: setup and secrets instructions.

## Task 1: Config and Environment

- [ ] Write tests in `internal/config/config_test.go` for loading YAML defaults and rejecting missing secrets.
- [ ] Run `go test ./internal/config` and confirm it fails because the package does not exist.
- [ ] Implement `internal/config/config.go` with `Load(path string)` and `LoadEnv()`.
- [ ] Add `go.mod` and `config.yaml`.
- [ ] Run `go test ./internal/config` and confirm it passes.

## Task 2: Recommendation Core

- [ ] Write tests for category detection in `internal/recommend/category_test.go`.
- [ ] Run `go test ./internal/recommend -run TestDetectCategories` and confirm it fails.
- [ ] Implement `internal/recommend/model.go` and `internal/recommend/category.go`.
- [ ] Run the category tests and confirm they pass.
- [ ] Write scoring tests in `internal/recommend/scoring_test.go` covering Go preference, activity, and recently recommended penalties.
- [ ] Run scoring tests and confirm they fail.
- [ ] Implement `internal/recommend/scoring.go`.
- [ ] Run scoring tests and confirm they pass.

## Task 3: History and Message Formatting

- [ ] Write history tests in `internal/recommend/history_test.go` for missing files, save/load, and recent duplicate detection.
- [ ] Run history tests and confirm they fail.
- [ ] Implement `internal/recommend/history.go`.
- [ ] Run history tests and confirm they pass.
- [ ] Write message formatting tests in `internal/recommend/message_test.go`.
- [ ] Run message tests and confirm they fail.
- [ ] Implement `internal/recommend/message.go`.
- [ ] Run message tests and confirm they pass.

## Task 4: External Clients

- [ ] Implement `internal/github/client.go` with focused search queries for Go, AI/agent, and DevOps/infra categories.
- [ ] Implement `internal/telegram/client.go` using `sendMessage`.
- [ ] Keep these clients thin and rely on core package tests for behavior.

## Task 5: CLI and Workflow

- [ ] Implement `cmd/recommend/main.go` to load config/env, fetch candidates, rank recommendations, send Telegram, and save history.
- [ ] Add `.github/workflows/daily.yml` with UTC cron `0 21 * * *`, `workflow_dispatch`, Go setup, tests, CLI execution, and history commit.
- [ ] Add `README.md` with GitHub Secrets and Telegram chat id setup instructions.
- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/recommend -dry-run` to verify local formatting without secrets.
- [ ] Commit the implementation.

## Self-Review

- Spec coverage: The plan covers scheduled GitHub Actions, public config, environment secrets, GitHub collection, scoring, history, Telegram delivery, tests, and documentation.
- Placeholder scan: No TBD/TODO placeholders remain in the plan.
- Type consistency: Package names and file paths are consistent across tasks.

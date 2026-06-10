# Obsidian Learning Route Bot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate a two-day GitHub project learning route, write it into the configured Obsidian vault, and send a Telegram reminder with Obsidian and GitHub jump links.

**Architecture:** Keep the existing GitHub search and ranking flow. Add focused packages for learning state and Obsidian Markdown rendering/writing, then update the CLI orchestration to either advance the active two-day route or create a new route. Keep Telegram as a reminder channel only.

**Tech Stack:** Go 1.22, standard library, existing GitHub REST client, existing Telegram Bot API client, YAML config, JSON state files, Markdown files in Obsidian.

---

## File Structure

- Modify `config.yaml`: add Obsidian, analysis, and LLM config defaults.
- Modify `internal/config/config.go`: load and default the new config sections.
- Modify `internal/config/config_test.go`: cover new config defaults and custom values.
- Create `internal/learning/state.go`: load/save active two-day learning route state.
- Create `internal/learning/state_test.go`: test missing file, save/load, route-day transitions.
- Create `internal/learning/plan.go`: build deterministic template learning plans from a selected recommendation.
- Create `internal/learning/plan_test.go`: test plan content and repo-name slug behavior.
- Create `internal/obsidian/writer.go`: create project folders, render Markdown, build Obsidian deep links.
- Create `internal/obsidian/writer_test.go`: test repo-only folder names, non-overwrite behavior, deep-link encoding.
- Modify `internal/telegram/client.go`: support optional URL buttons through `reply_markup`.
- Create `internal/telegram/client_test.go`: test outgoing JSON payload for URL buttons.
- Modify `cmd/recommend/main.go`: add `-obsidian` mode, state handling, Obsidian writing, and reminder formatting.
- Modify `.github/workflows/daily.yml`: schedule remote recommendation workflow at 11:00 Asia/Shanghai only if keeping GitHub Actions enabled for Telegram-only runs.
- Modify `README.md`: document local Obsidian usage, required permissions, and API key/provider options.

## Task 1: Config Sections

- [ ] Write failing tests in `internal/config/config_test.go` for Obsidian and analysis defaults:

```go
if cfg.Obsidian.VaultPath != "/Users/janeshirley/Documents/Obsidian Vault" { t.Fatalf(...) }
if cfg.Obsidian.ProjectDir != "GitHub Projects" { t.Fatalf(...) }
if cfg.Analysis.Provider != "template" { t.Fatalf(...) }
if cfg.Analysis.ScheduleTime != "11:00" { t.Fatalf(...) }
```

- [ ] Run `go test ./internal/config` and verify the test fails because the new fields do not exist.
- [ ] Add `ObsidianConfig`, `AnalysisConfig`, and `LLMConfig` to `internal/config/config.go`.
- [ ] Add defaults for vault path, project dir, `template` provider, and `11:00` schedule.
- [ ] Update `config.yaml` with the new sections.
- [ ] Run `go test ./internal/config` and verify it passes.

## Task 2: Learning State

- [ ] Write failing tests in `internal/learning/state_test.go` for:
  - missing state file returns empty state;
  - saving and loading active route state;
  - Day 1 advances to Day 2 when due;
  - Day 2 completion makes the next run eligible for a new project.
- [ ] Run `go test ./internal/learning` and verify it fails because the package does not exist.
- [ ] Implement `internal/learning/state.go` with `State`, `Route`, `LoadState`, `SaveState`, `NeedsNewProject`, and `CurrentDay`.
- [ ] Run `go test ./internal/learning` and verify it passes.

## Task 3: Template Learning Plan

- [ ] Write failing tests in `internal/learning/plan_test.go` for generating a two-day plan from a `recommend.Recommendation`.
- [ ] Verify the plan uses repo-only slug, such as `gin` for `gin-gonic/gin`.
- [ ] Run `go test ./internal/learning` and verify failure.
- [ ] Implement `internal/learning/plan.go` with `Plan`, `DayPlan`, `BuildTemplatePlan`, and `RepoSlug`.
- [ ] Run `go test ./internal/learning` and verify it passes.

## Task 4: Obsidian Writer

- [ ] Write failing tests in `internal/obsidian/writer_test.go` for:
  - writing `/vault/GitHub Projects/gin/gin.md`;
  - preserving an existing file;
  - generating `obsidian://open?vault=Obsidian%20Vault&file=GitHub%20Projects%2Fgin%2Fgin.md`.
- [ ] Run `go test ./internal/obsidian` and verify it fails because the package does not exist.
- [ ] Implement `internal/obsidian/writer.go` with `Writer`, `WritePlan`, `RenderMarkdown`, and `DeepLink`.
- [ ] Run `go test ./internal/obsidian` and verify it passes.

## Task 5: Telegram URL Buttons

- [ ] Write failing tests in `internal/telegram/client_test.go` using `httptest.Server` to verify `SendMessage` can send URL buttons.
- [ ] Run `go test ./internal/telegram` and verify failure.
- [ ] Add `Button` and `SendOptions` types, keep existing `SendMessage` behavior, and add `SendMessageWithOptions`.
- [ ] Run `go test ./internal/telegram` and verify it passes.

## Task 6: CLI Orchestration

- [ ] Write focused tests for helper functions in `cmd/recommend/main_test.go`: reminder text, state path defaults, and Obsidian mode behavior where possible without network.
- [ ] Run `go test ./cmd/recommend` and verify failure.
- [ ] Update `cmd/recommend/main.go`:
  - add `-obsidian`;
  - load `data/learning_state.json`;
  - generate a new plan when needed;
  - write Obsidian Markdown;
  - send Telegram reminder with Obsidian and GitHub URL buttons;
  - keep existing dry-run behavior.
- [ ] Run `go test ./cmd/recommend`.

## Task 7: Documentation and Workflow

- [ ] Update `README.md` with local Obsidian setup and the 11:00 schedule.
- [ ] Update `.github/workflows/daily.yml` cron to `0 3 * * *` only if the workflow remains responsible for Telegram-only reminders.
- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/recommend -dry-run -obsidian` to verify local formatting without secrets.

## Self-Review

- Spec coverage: the plan covers Obsidian folder naming, Markdown report generation, Telegram deep links, two-day state, 11:00 schedule documentation, and provider-neutral analysis configuration.
- Placeholder scan: no TBD or TODO placeholders are left.
- Type consistency: config, learning, obsidian, telegram, and CLI responsibilities are split into small packages with stable names.

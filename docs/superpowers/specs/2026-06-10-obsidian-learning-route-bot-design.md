# Obsidian Learning Route Bot Design

## Goal

Upgrade the current GitHub recommendation bot from a daily project link sender into a lightweight learning-route generator.

The bot should recommend one GitHub project at a time, generate a two-day learning route that covers the project's main learning value, write the report into the user's Obsidian vault, and send a Telegram reminder with jump links.

The first version should stay simple:

- No long-running Telegram callback service.
- No web UI.
- Obsidian checkboxes are the check-in mechanism.
- Telegram only reminds the user and links to the generated note.

## Schedule

Generate the Obsidian learning report every day at 11:00 Asia/Shanghai.

The product behavior is still project-based rather than day-based:

- If there is no active project, select a new project and generate a two-day route.
- If the active project is on Day 1, remind the user to complete Day 1.
- If the active project is on Day 2, remind the user to complete Day 2.
- After Day 2 is due, the next scheduled run can select a new project.

This keeps the workflow predictable: one project gets a two-day learning window, then the system moves on.

## Obsidian Output

The Obsidian vault path is:

```yaml
obsidian:
  vault_path: "/Users/janeshirley/Documents/Obsidian Vault"
  project_dir: "GitHub Projects"
```

Each project gets its own folder under the configured project directory.

Example for `gin-gonic/gin`:

```text
/Users/janeshirley/Documents/Obsidian Vault/GitHub Projects/
  gin/
    gin.md
```

Obsidian folder and Markdown file names use only the repo name, such as `gin`. The full `owner/repo` name is written into the Markdown metadata and body links.

The Markdown file should include:

- Project metadata and GitHub link.
- Obsidian internal links where useful.
- Why the project is worth learning.
- Architecture or framework design highlights.
- Business or engineering hard points.
- Day 1 checklist.
- Day 2 checklist.
- Review questions.
- A free-form notes section for the user.

## Telegram Reminder

Telegram remains a reminder channel, not the source of truth.

The reminder should include:

- Project name.
- Current day in the route, such as Day 1 or Day 2.
- GitHub URL.
- Obsidian deep link.
- Relative Obsidian path as fallback text.

Example:

```text
今日学习项目：gin-gonic/gin

路线已写入 Obsidian：
GitHub Projects/gin/gin.md

打开 Obsidian：
obsidian://open?vault=Obsidian%20Vault&file=GitHub%20Projects%2Fgin%2Fgin.md

GitHub：
https://github.com/gin-gonic/gin

今天完成 Day 1。
```

If Telegram URL buttons are added, they should only be URL buttons:

- Open Obsidian.
- Open GitHub.

## Analysis Backend

The analysis backend should be replaceable.

Initial providers:

1. `codex`: use Codex non-interactive or automation-driven analysis when running locally.
2. `openai_compatible`: call an OpenAI-compatible API endpoint using a configurable base URL, model, and API key environment variable.
3. `template`: fallback deterministic template when no model backend is configured.

Example config:

```yaml
analysis:
  provider: "codex"
  schedule_time: "11:00"

llm:
  provider: "openai_compatible"
  base_url: "https://api.example.com/v1"
  model: "your-model"
  api_key_env: "LLM_API_KEY"
```

The first implementation can start with `template` plus one model provider. The code should avoid hard-coding OpenAI-specific assumptions so other OpenAI-compatible services can be used later.

Codex is useful for local, personal automation because it can inspect repository context and produce high-quality explanations. It is less portable than a plain API backend because Codex automation depends on the local Codex app or CLI environment, local permissions, and scheduled execution on the user's machine.

## GitHub Project Context Collection

The current bot already searches and ranks repositories. The learning-route version should keep that flow and enrich the selected project before generating a report.

Minimum context to collect:

- Repository metadata from the current search result.
- README content.
- Top-level directory tree.
- Key dependency files when present, such as `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`.
- `docs`, `examples`, `cmd`, `internal`, `pkg`, `test`, and similar learning-relevant directories when available.

The first version should not clone full repositories by default. Prefer GitHub API content calls and keep the context bounded.

## State

The existing `data/recommendations.json` only tracks recommendation history. The learning-route version needs additional state.

Suggested state file:

```text
data/learning_state.json
```

It should track:

- Active project full name.
- Project slug, such as `gin`.
- Obsidian file path.
- Generated date.
- Current route day.
- Due dates for Day 1 and Day 2.
- Whether the project is complete, skipped, or active.

The user checks off tasks in Obsidian manually. The bot does not need to parse checkbox completion in the first version.

## Error Handling

- If Obsidian vault path is missing, fail with a clear error.
- If the Obsidian file already exists for a project, do not overwrite user notes. Append a small update only if explicitly configured later.
- If model analysis fails, write a template-based note and mention that model analysis failed.
- If Telegram send fails after the Obsidian note is written, keep the note and return a non-zero exit code so the scheduled run is visible as failed.
- If no new project is available, send a short Telegram message and do not create an empty Obsidian note.

## Out of Scope

- Long-running Telegram bot webhook service.
- Obsidian plugin development.
- Web dashboard.
- Automatic parsing of Obsidian checkbox completion.
- Automatic browser control of ChatGPT web UI.

## Success Criteria

- A scheduled run at 11:00 Asia/Shanghai can generate or advance a two-day learning route.
- Each selected project creates one Obsidian folder and one Markdown report.
- The report contains enough guidance to know what to read, what to build, and what to reflect on over two days.
- Telegram sends a reminder with an Obsidian deep link and GitHub link.
- The implementation remains small and does not require a persistent server.

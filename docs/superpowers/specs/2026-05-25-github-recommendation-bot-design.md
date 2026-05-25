# GitHub Recommendation Bot Design

## Goal

Build a daily GitHub recommendation bot that sends high-quality technical project recommendations to a Telegram private chat. The repository will be public, so all sensitive values must be provided through environment variables or GitHub Actions secrets.

The bot should prefer Go projects, but it should not be limited to Go. It should also recommend high-quality projects related to AI, agent development, and DevOps/infra when they have strong learning value.

## Delivery Model

The bot will be implemented as a Go CLI and run from GitHub Actions.

The scheduled workflow will run every day at 13:00 Asia/Shanghai. To avoid relying on timezone-specific workflow syntax, the workflow should use UTC cron:

```yaml
cron: "0 5 * * *"
```

The workflow should also support `workflow_dispatch` so recommendations can be tested manually from the GitHub Actions UI.

## Configuration

Public, non-sensitive settings live in `config.yaml`.

Example:

```yaml
recommend:
  count: 5
  min_stars: 500
  lookback_days: 30
  exclude_repos: []
  categories:
    - go
    - ai_agent
    - devops_infra
telegram:
  parse_mode: Markdown
```

Sensitive settings must be read from environment variables:

- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`
- `GITHUB_TOKEN`

In GitHub Actions these values should be configured as repository secrets. The token and chat id must never be committed to the public repository.

## Recommendation Scope

Recommendations should be high-quality technical projects, with this priority order:

1. Go-related projects: Go language projects, Go tooling, backend frameworks, infra tooling, and repositories tagged with `golang`.
2. AI and agent development: LLM apps, agent frameworks, RAG tooling, MCP tooling, evaluation frameworks, prompt/workflow automation, and developer tools for AI systems.
3. DevOps and infrastructure: Kubernetes, observability, CI/CD, deployment, databases, security tools, and operational automation.
4. Other high-quality engineering projects: allowed only when they have unusually strong learning value and meet stricter quality thresholds.

The bot should label each recommendation with one or more categories so the Telegram message makes the recommendation reason clear.

## GitHub Data Collection

The CLI will query GitHub Search API with multiple focused searches instead of one broad query.

Initial query families:

- Go: `language:Go stars:>500`, `topic:golang stars:>500`, and recently pushed Go repositories.
- AI/agent: queries around `topic:llm`, `topic:agents`, `topic:rag`, `topic:mcp`, `topic:openai`, and related terms with star thresholds.
- DevOps/infra: queries around `topic:kubernetes`, `topic:observability`, `topic:devops`, `topic:ci-cd`, `topic:terraform`, and similar infra terms.

Archived repositories and obvious forks should be filtered out where GitHub metadata allows it.

## Scoring

Each candidate gets a deterministic score based on:

- Popularity: stars, forks, and watchers.
- Activity: recent push time and release freshness when available.
- Health: archived/fork status, issue ratio, license presence, homepage presence, topics, and description quality.
- Category fit: Go gets the highest preference weight, AI/agent and DevOps/infra receive strong secondary weights, and general projects require higher quality.
- Recommendation history: recently recommended repositories should be skipped or heavily penalized.

The scoring does not need to perfectly model quality. It should be transparent, deterministic, and easy to tune from `config.yaml`.

## State

GitHub Actions runners are ephemeral, so recommendation history should be stored in the repository.

Use `data/recommendations.json` to store public recommendation state:

```json
{
  "items": [
    {
      "full_name": "owner/repo",
      "recommended_at": "2026-05-25",
      "score": 123.45,
      "categories": ["go", "devops_infra"]
    }
  ]
}
```

After a successful recommendation run, the workflow should commit changes to this file back to the default branch. The file must not contain tokens, chat ids, or private notes.

## Telegram Message

The Telegram message should be concise and useful in Chinese.

Example structure:

```text
今日推荐：5 个值得学习的 GitHub 项目

1. owner/repo
分类：Go / Infra
推荐理由：...
亮点：...
链接：https://github.com/owner/repo
```

The message should include the repository name, categories, short recommendation reason, key metrics, and link.

## Error Handling

- Missing environment variables should fail fast with a clear error.
- GitHub API errors or rate limits should fail the run instead of sending incomplete recommendations.
- Telegram send failure should return a non-zero exit code so the GitHub Actions run is marked failed.
- If too few candidates remain after filtering, the CLI may relax soft thresholds, but must keep the category and quality rules.
- If the history file is missing, the CLI should create it.

## Tests

Unit tests should cover:

- Configuration loading and environment validation.
- Candidate category detection.
- Candidate scoring and sorting.
- History-based de-duplication.
- Telegram message formatting.

GitHub and Telegram API calls should be behind small interfaces so tests can use mock clients without network access.

## GitHub Actions Workflow

The workflow should:

1. Check out the repository.
2. Set up Go.
3. Run `go test ./...`.
4. Run the recommendation CLI with secrets injected as environment variables.
5. Commit `data/recommendations.json` back to the default branch only when it changed.

The workflow should request only the permissions it needs, including `contents: write` for committing the public history file.

## Out of Scope

The first version will not implement a long-running Telegram command bot, subscription management, a web UI, or per-user preference editing. These can be added later after the scheduled daily recommendation flow is stable.

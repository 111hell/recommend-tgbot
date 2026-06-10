# GitHub Recommendation Bot

Daily GitHub project recommendations delivered to a Telegram private chat.

The bot prefers Go projects, but it can also recommend high-quality AI/agent and DevOps/infra projects with strong learning value.

## How It Runs

GitHub Actions runs the Go CLI every day at 11:00 Asia/Shanghai for the Telegram-only recommendation flow. The workflow uses UTC cron:

```yaml
cron: "0 3 * * *"
```

Manual runs are available from the GitHub Actions UI through `workflow_dispatch`.

## Obsidian Learning Route Mode

The bot can also generate a two-day learning route for one project and write it to Obsidian:

```bash
GITHUB_TOKEN=your_github_token \
TELEGRAM_BOT_TOKEN=your_telegram_bot_token \
TELEGRAM_CHAT_ID=your_chat_id \
go run ./cmd/recommend -obsidian
```

The Obsidian mode writes to:

```text
/Users/janeshirley/Documents/Obsidian Vault/GitHub Projects/<repo>/<repo>.md
```

For example, `gin-gonic/gin` becomes:

```text
/Users/janeshirley/Documents/Obsidian Vault/GitHub Projects/gin/gin.md
```

Telegram receives a reminder with:

- the project name;
- the GitHub link;
- the Obsidian relative path;
- an `obsidian://open` deep link.

Use a local scheduler such as Codex Automation, cron, or launchd to run Obsidian mode at 11:00 on your machine. GitHub-hosted Actions cannot write to your local Obsidian vault.

Preview without secrets or file writes:

```bash
go run ./cmd/recommend -dry-run -obsidian
```

## Configuration

Public recommendation settings live in `config.yaml`.

Do not put secrets in `config.yaml`.

Required GitHub repository secrets:

- `TELEGRAM_BOT_TOKEN`: token from BotFather.
- `TELEGRAM_CHAT_ID`: your private chat id with the bot.

The workflow uses GitHub's built-in `GITHUB_TOKEN` for GitHub API calls and for committing public recommendation history.

## Telegram Chat ID

1. Create a bot with BotFather and copy the bot token.
2. Send any message to your bot from your Telegram account.
3. Open `https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/getUpdates`.
4. Find `message.chat.id` and save it as `TELEGRAM_CHAT_ID`.

## Local Development

Run tests:

```bash
go test ./...
```

Run a dry run without Telegram secrets:

```bash
GITHUB_TOKEN=your_github_token go run ./cmd/recommend -dry-run
```

The first real workflow run creates `data/recommendations.json` and commits it back to the repository.

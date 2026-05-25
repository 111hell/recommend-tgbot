# GitHub Recommendation Bot

Daily GitHub project recommendations delivered to a Telegram private chat.

The bot prefers Go projects, but it can also recommend high-quality AI/agent and DevOps/infra projects with strong learning value.

## How It Runs

GitHub Actions runs the Go CLI every day at 13:00 Asia/Shanghai. The workflow uses UTC cron:

```yaml
cron: "0 5 * * *"
```

Manual runs are available from the GitHub Actions UI through `workflow_dispatch`.

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

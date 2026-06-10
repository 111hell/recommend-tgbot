# Obsidian 学习路线 Bot 设计

## 目标

把当前的 GitHub 推荐 bot 从“每天发一个项目链接”，升级成一个轻量的学习路线生成器。

Bot 每次推荐一个 GitHub 项目，为这个项目生成一份两天可以完成的学习路线，把报告写入 Obsidian Vault，并通过 Telegram 发送带跳转链接的提醒。

第一版保持精简：

- 不做长期运行的 Telegram callback 服务。
- 不做 Web UI。
- 打卡直接用 Obsidian 里的 checkbox。
- Telegram 只负责提醒和跳转。

## 调度

每天上午 11:00，按 Asia/Shanghai 时区生成或推进 Obsidian 学习报告。

产品行为以“项目”为单位，而不是单纯按天推链接：

- 如果当前没有学习中的项目，选择一个新项目，生成两天学习路线。
- 如果当前项目处于 Day 1，提醒完成 Day 1。
- 如果当前项目处于 Day 2，提醒完成 Day 2。
- Day 2 到期后，下一次定时任务可以选择新项目。

这样每个项目都有一个清晰的两天学习窗口，两天后再进入下一个项目。

## Obsidian 输出

Obsidian Vault 路径固定为：

```yaml
obsidian:
  vault_path: "/Users/janeshirley/Documents/Obsidian Vault"
  project_dir: "GitHub Projects"
```

每个项目在 `GitHub Projects` 下面单独创建一个文件夹。

例如 `gin-gonic/gin`：

```text
/Users/janeshirley/Documents/Obsidian Vault/GitHub Projects/
  gin/
    gin.md
```

Obsidian 文件夹名和 Markdown 文件名只使用 repo 名，例如 `gin`。完整的 `owner/repo` 写入 Markdown 报告元信息和正文链接里。

每个项目的 Markdown 报告包含：

- 项目元信息和 GitHub 链接。
- 必要的 Obsidian 内部链接。
- 为什么这个项目值得学习。
- 架构或框架设计亮点。
- 业务或工程上的重难点。
- Day 1 学习 checklist。
- Day 2 学习 checklist。
- 复盘问题。
- 用户自由记录笔记的区域。

## Telegram 提醒

Telegram 是提醒渠道，不是学习内容的主存储。

提醒内容包含：

- 项目名。
- 当前学习进度，例如 Day 1 或 Day 2。
- GitHub 链接。
- Obsidian 深链。
- Obsidian 相对路径，作为深链失效时的备用文本。

示例：

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

如果之后给 Telegram 消息加按钮，第一版只加 URL 按钮：

- 打开 Obsidian。
- 打开 GitHub。

## 分析后端

项目分析后端要做成可替换的，不绑定某一家模型服务。

初始支持三类：

1. `codex`：本地运行时，可以用 Codex non-interactive 或 Codex Automation 做项目分析。
2. `openai_compatible`：调用兼容 OpenAI API 格式的模型服务，通过 base URL、model、环境变量配置 API key。
3. `template`：没有配置模型时的兜底模板，生成确定性的基础学习报告。

示例配置：

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

第一版可以先实现 `template`，再实现一种模型后端。代码结构上不要写死 OpenAI，这样之后可以接其他兼容 OpenAI API 的服务。

Codex 适合本地个人自动化，因为它可以读取项目上下文，解释质量也可能更高。但它比普通 API 后端更依赖本机环境：需要 Codex app 或 CLI、需要本机权限、需要你的机器在定时任务运行时可用。所以它适合个人工作流，不适合作为可移植开源 bot 的唯一后端。

## GitHub 项目上下文收集

当前 bot 已经有搜索和排序逻辑。学习路线版本应该保留这条链路，在选出项目后再收集更多上下文。

第一版至少收集：

- 当前搜索结果里的仓库元信息。
- README 内容。
- 顶层目录结构。
- 关键依赖文件，例如 `go.mod`、`package.json`、`pyproject.toml`、`Cargo.toml`。
- 有学习价值的目录，例如 `docs`、`examples`、`cmd`、`internal`、`pkg`、`test`。

第一版默认不 clone 完整仓库。优先用 GitHub API 读取 README、目录树和少量关键文件，控制上下文大小。

## 状态

现有的 `data/recommendations.json` 只记录推荐历史。学习路线版本需要新增学习状态。

建议新增：

```text
data/learning_state.json
```

它记录：

- 当前学习项目的 full name。
- 项目 slug，例如 `gin`。
- Obsidian 文件路径。
- 报告生成日期。
- 当前学习日，例如 Day 1 或 Day 2。
- Day 1 和 Day 2 的截止日期。
- 项目状态：active、complete 或 skipped。

第一版不解析 Obsidian checkbox 是否完成。你在 Obsidian 里手动勾选即可。

## 错误处理

- 如果 Obsidian Vault 路径不存在，直接失败，并给出清晰错误。
- 如果项目 Markdown 已经存在，不覆盖用户笔记。后续可以再加显式配置决定是否追加更新。
- 如果模型分析失败，写入模板版报告，并在报告里说明模型分析失败。
- 如果 Obsidian 报告已经写入，但 Telegram 发送失败，保留报告，同时让命令返回非零退出码，方便发现定时任务失败。
- 如果没有找到新项目，发送一条简短 Telegram 提醒，不创建空报告。

## 不做的事

第一版明确不做：

- 长期运行的 Telegram webhook 服务。
- Obsidian 插件开发。
- Web dashboard。
- 自动解析 Obsidian checkbox 完成状态。
- 自动控制 ChatGPT 网页 UI。

## 成功标准

- 每天上午 11:00 Asia/Shanghai 可以生成或推进两天学习路线。
- 每个项目在 Obsidian 里有一个文件夹和一个 Markdown 报告。
- 报告能指导你两天内知道该读什么、做什么、复盘什么。
- Telegram 提醒包含 Obsidian 深链和 GitHub 链接。
- 整体实现保持小而清晰，不需要长期运行的服务器。

# rook

**A local mission-control for your AI coding agents.**

rook is a self-hosted, single-user dashboard that watches every AI coding agent
running on your machine — Claude Code, Codex, Aider, Gemini — and pulls you in the
moment one needs you. Approve a permission prompt, reply, jump into the agent's live
terminal, review its diff, read the PR it's working on, or launch a new agent from a
GitHub/Linear/Jira ticket — all from one screen. Everything runs on `localhost`, and
GitHub access is read-only unless you explicitly turn on write actions.

![The Operator console](docs/img/operator.png)

<sub>The Operator console: a live agent roster on the left, and the selected agent's workspace on the right — Overview, Terminal, Diff, Trace, Files, and the PR/issue it's working on.</sub>

---

## Why rook

If you run more than one coding agent at a time, you lose track. One is blocked on a
permission prompt, another finished ten minutes ago, a third is burning tokens on the
wrong thing — and you're cycling through terminal tabs to find out. rook is one pane of
glass over all of them:

- **Never miss a waiting agent** — a live "needs you" count, and Allow / Deny / reply
  right from the workspace (or the agent's real terminal, embedded via tmux).
- **See what an agent is actually doing** — a context-window gauge (how full it is),
  tool-usage breakdown, an execution-trace waterfall, and its live diff.
- **Full PR/issue context, in-app** — the description, commits, linked issues, reviews,
  and comments for the PR an agent is reviewing, pulled live from GitHub.
- **Start work without typing paths** — hand a GitHub/Linear/Jira ticket to an agent;
  rook fetches the ticket, writes the task, and auto-resolves the local checkout.
- **Know where your tokens go** — 5-hour and 7-day windows with input/cache breakdown,
  cost by model and by project, and a 30-day activity trend.
- **Stay in control** — a searchable audit trail of every command (with risky-command
  flags), a git-worktree manager, and a dev-server panel to kill stray `npm run dev`s.
- **Dark and light themes**, keyboard-driven, command palette (⌘K) for everything.

## Screenshots

| Insights | Board |
|---|---|
| ![Insights](docs/img/insights.png) | ![Board](docs/img/board.png) |
| Usage, cost by model & project, 30-day trend | Every agent by live state |

| Inline diff review | PR / issue context |
|---|---|
| ![Diff](docs/img/diff.png) | ![PR context](docs/img/pr-context.png) |
| File tree, split/unified, comments → agent | Description, commits, linked issues, comments |

![Light theme](docs/img/light.png)
<sub>Light theme — toggle it from the rail; your choice is remembered.</sub>

## Supported agents

| Agent | Monitoring | Control (tmux) |
|-------|-----------|----------------|
| Claude Code | stable | yes |
| Codex | beta | yes |
| Aider | beta | yes |
| Gemini | beta | yes |

Monitoring works for any agent that writes session transcripts to disk. Control
(Allow/Deny, live terminal, spawn) requires the agent to run inside a `tmux` session.

## Requirements

- **Go** — a recent toolchain to build (see the `go` directive in [`go.mod`](go.mod)).
- **tmux** *(optional, recommended)* — for control features (terminal, Allow/Deny,
  spawn). `brew install tmux`.
- **gh CLI** *(optional)* — for the GitHub view, PR/issue context, clone, and PR
  create/merge. `brew install gh && gh auth login`.
- The agents themselves (e.g. [Claude Code](https://claude.com/claude-code)) writing
  their session data under `~/.claude`.

## Quick start

```bash
git clone https://github.com/PiyushSingh-ZS/rook.git
cd rook

make run            # builds ./rook and starts it (loopback, port 7480)
# or: make build && ./rook

open http://127.0.0.1:7480
```

That's it — rook auto-discovers the agents already running on your machine.

> **New here? Read the [Setup & Usage Guide](docs/USAGE.md)** — install, wiring up tmux
> + gh, and a walkthrough of every part of the UI.

### Control + notifications

- **Control an agent:** launch it inside tmux so rook can drive it —
  `tmux new -s my-task claude` — or use the **+ / Launch agent** button in rook.
- **Approve dangerous commands from rook:** Settings → install the Claude Code hooks
  bridge, then turn on the destructive-command gate.
- **Phone push:** Settings → paste an [ntfy](https://ntfy.sh) topic URL. Desktop
  notifications are on by default (`--notify`).

## Configuration

**Flags:**

| Flag | Default | Purpose |
|------|---------|---------|
| `--addr` | `127.0.0.1:7480` | Listen address (loopback only by default). |
| `--notify` | `true` | Desktop notification when an agent starts waiting. |
| `--token` | *(empty)* | Require this token for non-loopback clients (e.g. over Tailscale). |

**Environment:**

- `CLAUDE_CONFIG_DIR` — point rook at a non-default Claude config directory.

**In-app (Settings):** automation (hooks gate, auto-review, auto-verify, allow write
actions), notifications (ntfy / Slack / Discord), editor, and Linear/Jira/summary
config. Settings persist to `~/.rook/config.json`.

## Security

- Binds to `127.0.0.1` only; non-loopback requests are rejected unless you set
  `--token`.
- **GitHub is read-only by default.** PR create/merge are off until you enable *Allow
  write actions* in Settings.
- Reads local agent data from `~/.claude` (and other agent dirs) and your project
  directories. Nothing leaves your machine.
- The only state-changing actions are ones you trigger: sending keystrokes to a tmux
  pane you own, spawning an agent, stopping a dev server, deleting a worktree.

## How it works

rook is a single Go binary (built on [GoFr](https://gofr.dev)) with an embedded,
zero-build web UI. It polls agent session files on disk, maps sessions to tmux panes
via the process tree, parses transcripts for tokens/tool-calls/changed-files, and
estimates cost from token counts. The browser polls `/api/state` every 2s.

```
~/.claude/**/*.jsonl        → sessions, transcripts, tool calls, files changed, context
tmux capture-pane/send-keys → live terminal + Allow/Deny/reply/spawn
gh (read-only by default)   → repos, issues, PRs, PR/issue context
git remotes                 → auto-resolve a repo's local checkout
~/.rook/                    → config.json, rook.db, worktrees/
```

For a deeper map of the codebase — and to have an AI agent work on rook — see
[AGENTS.md](AGENTS.md).

## Contributing

Contributions welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for dev setup, the
build/test/lint commands, and the frontend architecture (how to add a new view).

## License

[MIT](LICENSE).

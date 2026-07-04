<div align="center">

# relay

**The first MCP server for product launches.**

[![CI](https://img.shields.io/github/actions/workflow/status/valtors/relay/ci.yml?style=flat&label=ci)](https://github.com/valtors/relay/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/valtors/relay?style=flat)](https://github.com/valtors/relay/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/valtors/relay)](https://goreportcard.com/report/github.com/valtors/relay)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/valtors/relay)](go.mod)
[![Contributors welcome](https://img.shields.io/badge/contributors-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Good first issue](https://img.shields.io/github/issues/valtors/relay/good%20first%20issue?label=good%20first%20issue)](https://github.com/valtors/relay/labels/good%20first%20issue)

A multi-agent pipeline that takes your product from idea to launch plan.
Five stages. Human checkpoints between each. One Go binary. Runs inside any MCP client.

[Quick Start](#quick-start) · [How It Works](#how-it-works) · [Install](#install) · [Contributing](#contributing)

</div>

---

## New here?

Want the easiest first contribution?

- Read [docs/ADDING_A_TOOL.md](docs/ADDING_A_TOOL.md)
- Browse [good first issues](https://github.com/valtors/relay/labels/good%20first%20issue)
- Ask questions in [GitHub Discussions](https://github.com/valtors/relay/discussions)

We try to reply to contributor questions and PRs within 3 business days.

---

## What is Relay?

Relay is an MCP server written in Go. You wire it into your editor (Claude Code, Cursor, VS Code Copilot, Windsurf) and it runs a structured, resumable product launch workflow:

```
Brief → PM → Research → Brand → UX → GTM → Launch Plan
         ↑       ↑        ↑      ↑      ↑
      approve  approve  approve approve approve
```

Every stage produces a real artifact. Every stage pauses for your approval before continuing. Your data stays on your machine.

---

## Why Relay?

You shipped a product. Now what? Who do you tell? What do you say? In what order?

Most founders ask ChatGPT, get generic slop, post once on X, get 4 likes, and give up. Relay forces a **plan before any copy** and grounds everything in your specific audience, positioning, and channels.

| Problem | How Relay Solves It |
|---|---|
| ChatGPT forgets your context every session | Relay persists state as JSON on disk |
| No structure to marketing work | Five opinionated stages with real deliverables |
| Generic AI copy that sounds like everyone else | Each stage builds on the last, grounded in YOUR product |
| Launch day comes and goes with no plan | GTM stage produces a sequenced, multi-channel playbook |
| Tools cost $100+/mo and need a team | Single binary, MIT license, bring your own API key |

---

## How It Works

### The 5-Stage Pipeline

| Stage | Agent | Artifact |
|---|---|---|
| **PM** | Product Manager | `prd.md` - problem, users, scope, success metrics |
| **Research** | Researcher | `research.md` - market scan, competitors, sources |
| **Brand** | Brand Strategist | `brand.md` - positioning, voice, messaging |
| **UX** | UX Architect | `ux.md` - user flows, screen inventory, key states |
| **GTM** | Go-to-Market | `gtm.md` - launch checklist, channels, copy drafts |

### The 8 MCP Tools

```
start_workflow    Kick off a new run from a one-line idea
get_state         Current phase, status, last artifact
list_phases       Pipeline overview with completion status
start_phase       Begin a specific phase (idempotent)
get_artifact      Return the rendered artifact for a phase
approve           Advance past the current checkpoint
revise            Loop the current phase with feedback
abort             Clean shutdown, preserve state
```

---

## Quick Start

### Prerequisites

- Any MCP-compatible editor (Claude Code, Cursor, VS Code Copilot, Windsurf)
- An Anthropic API key

### Install

**Option A: Download binary**

```bash
# macOS / Linux
curl -fsSL https://github.com/valtors/relay/releases/latest/download/relay_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv relay /usr/local/bin/
```

```powershell
# Windows
gh release download --repo valtors/relay --pattern "*windows_amd64.zip"
```

**Option B: Build from source**

```bash
go install github.com/valtors/relay@latest
```

### Configure Your Editor

Add to your MCP config (`.mcp.json` or editor settings):

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-..."
      }
    }
  }
}
```

### Run

```
> start_workflow "A CLI tool that helps developers write better commit messages"
```

Relay kicks off the PM stage. Review the PRD it produces, then `approve` to advance or `revise "focus more on teams"` to iterate.

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Your Editor (Claude Code / Cursor / Copilot / etc) │
│      │                                               │
│      │  stdio (JSON-RPC, MCP)                        │
│      ▼                                               │
│  ┌───────────────────────────────────────────┐      │
│  │  relay  (single Go binary, ~12 MB)        │      │
│  │  ─────────────────────────────────────    │      │
│  │  MCP server       (stdio transport)       │      │
│  │  Workflow engine   (5 stages, resumable)   │      │
│  │  Checkpoints      (JSON on disk)          │      │
│  │  LLM client       (Anthropic, your key)   │      │
│  └───────────────────────────────────────────┘      │
│      │                                               │
│      ▼                                               │
│  ./.relay/                                           │
│      ├── state.json                                  │
│      ├── artifacts/{stage}/...                       │
│      └── logs/                                       │
└─────────────────────────────────────────────────────┘
```

---

## Design Principles

| Principle | In Practice |
|---|---|
| **Human-in-the-loop** | Every stage ends with an explicit checkpoint. The agent never advances without your approval. |
| **Local-first** | One binary. No cloud. No telemetry. Your API keys and artifacts never leave your machine. |
| **Resumable** | State is plain JSON on disk. Crashes, reboots, Ctrl-C don't lose work. Pick up days later. |
| **MCP-native** | Speaks the Model Context Protocol. Any MCP-aware editor just works. |
| **Boring tech** | Go stdlib. JSON files. No frameworks. Fewer moving parts. |

---

## Stack

| Layer | Choice |
|---|---|
| Language | Go |
| Transport | MCP over stdio |
| LLM | Anthropic API (bring your own key) |
| State | JSON files on disk |
| Build | GitHub Actions + GoReleaser |
| Platforms | darwin/linux/windows, amd64/arm64 |

---

## Contributing

We welcome contributions. Whether it is fixing a bug, improving prompts, adding a new stage, or writing docs.

### Getting Started

1. Fork the repo
2. Clone your fork
3. Run tests: `go test ./...`
4. Make your changes
5. Open a PR with a clear description of what changed and why

### Areas Where Help is Needed

- Prompt engineering for each agent stage
- Additional output formats (JSON, HTML)
- New stages (Analytics, Email sequences, Content calendar)
- Editor-specific integration guides
- Testing across different MCP clients

See [open issues](https://github.com/valtors/relay/issues) for specific tasks.

---

## Roadmap

- [x] Core 5-stage pipeline
- [x] Checkpoint approval system
- [x] Resumable state
- [x] Cross-platform binaries
- [ ] GoReleaser integration
- [ ] Web UI (relay-web)
- [ ] Plugin system for custom stages
- [ ] Multi-LLM support (OpenAI, local models via Ollama)
- [ ] Template library for common product types
- [ ] Post-launch persistence (weekly content batches)

---

## License

MIT - see [LICENSE](LICENSE) for details.

---

<div align="center">

**[Star on GitHub](https://github.com/valtors/relay)** · **[Report a Bug](https://github.com/valtors/relay/issues)** · **[Request a Feature](https://github.com/valtors/relay/issues)**

Built by [Tamish](https://github.com/tamish560) at [Valtors](https://github.com/valtors)

</div>

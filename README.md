# relay

A Model Context Protocol (MCP) server in Go that orchestrates a 5-stage,
multi-agent product launch pipeline (PM → Research → Brand → UX → GTM)
with mandatory human checkpoints between every stage.

Single binary. Stdio transport. Survives crashes. Works in Claude Code,
VS Code Copilot, Cursor, Windsurf, GitHub Copilot CLI.

[![ci](https://github.com/valtors/relay/actions/workflows/ci.yml/badge.svg)](https://github.com/valtors/relay/actions/workflows/ci.yml)
[![release](https://github.com/valtors/relay/actions/workflows/release.yml/badge.svg)](https://github.com/valtors/relay/actions/workflows/release.yml)

---

## 🚀 Quickstart tour (5 minutes)

> Goal: go from "fresh clone" to "Claude Code is running the full PM→GTM pipeline on a sample brief".

### 1. Get the binary

This repo is private. The official distribution channel is GitHub Releases,
downloaded via the [GitHub CLI](https://cli.github.com/) (`gh`) so you
authenticate once and never paste a token into a curl command.

**Option A — download a release with `gh`** (recommended):

```bash
# one-time: install + auth
gh auth login

# pick the binary for your OS — only one of these:
gh release download --repo valtors/relay --pattern "RELAY-darwin-arm64*"        # macOS Apple Silicon
gh release download --repo valtors/relay --pattern "RELAY-darwin-amd64*"        # macOS Intel
gh release download --repo valtors/relay --pattern "RELAY-linux-amd64*"         # Linux x86_64
gh release download --repo valtors/relay --pattern "RELAY-linux-arm64*"         # Linux arm64
gh release download --repo valtors/relay --pattern "RELAY-windows-amd64.exe*"   # Windows

# verify checksum, then install onto your PATH
# macOS / Linux:
sha256sum --check RELAY-*.sha256
chmod +x RELAY-* && sudo mv RELAY-* /usr/local/bin/relay

# Windows PowerShell:
$expected = (Get-Content RELAY-windows-amd64.exe.sha256).Split(" ")[0]
$actual   = (Get-FileHash RELAY-windows-amd64.exe -Algorithm SHA256).Hash.ToLower()
if ($expected -ne $actual) { throw "checksum mismatch" }
Move-Item RELAY-windows-amd64.exe "$env:USERPROFILE\bin\relay.exe"
```

**Option B — build from source:**

```bash
gh repo clone valtors/relay
cd relay
go install .   # binary lands in $(go env GOPATH)/bin
```

### 2. Set your Anthropic API key

```bash
export ANTHROPIC_API_KEY=sk-ant-...     # macOS / Linux
$env:ANTHROPIC_API_KEY = "sk-ant-..."   # Windows PowerShell
```

### 3. Make a workspace + brief

```bash
mkdir my-launch && cd my-launch
cat > product_brief.md <<'EOF'
# Product Brief: <your product name>

## What
One paragraph: what is it.

## Why
What pain it solves and why now.

## Target user
Who specifically, and what they currently do instead.

## Constraints
Budget, team size, ship date, business model.

## Success metric
One measurable goal for the first 90 days.
EOF
```

### 4. Wire it into your MCP client

Drop a `.mcp.json` in the workspace (Claude Code, Cursor, Windsurf all read this):

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay",
      "env": { "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}" }
    }
  }
}
```

Then launch your MCP client from the workspace dir:

```bash
claude         # or: cursor . / code . / copilot
```

Approve the `relay` server when prompted.

### 5. Run the pipeline

In your client's chat, ask:

> Use the `run_workflow` tool with `brief_path: "./product_brief.md"`.

The server will execute PM → Research → Brand → UX → GTM and pause at each
of four checkpoints (`H1`–`H4`). Reply `approve` to advance, `iterate <feedback>`
to redo a stage with notes, or `skip` to fast-forward.

When done you'll find every artifact in `./output/` and the stitched plan at
`./output/final_product_plan.md`.

> 💡 If the process crashes mid-pipeline, just call `run_workflow` again with
> the same brief — completed stages are skipped via `output/.session.meta.json`.

---

## Pipeline

```
product_brief.md
        │
        ▼
   pm_plan ─────────► output/pm_brief_for_agent1.md
        │
        ▼
 run_research ─────► output/01_research.md       (web search enabled)
        │
        ▼  H1: human approves or iterates with notes
 run_brand ────────► output/02_brand_messaging.md
        │
        ▼  H2
 run_ux ───────────► output/03_ux.md
        │
        ▼  H3
 run_gtm ──────────► output/04_go_to_market.md   (4a + 4b parallel)
        │
        ▼  H4
 assemble_plan ───► output/final_product_plan.md
```

Each `H*` checkpoint writes a reviewable `output/checkpoint_H*.md` file and
blocks on stdin. The human sends `approve` to advance or `iterate <notes>`
to re-run that stage with the notes injected into the agent's context.

## Tools (8)

| Tool | Purpose |
|------|---------|
| `run_workflow` | Master orchestrator — runs the entire pipeline with crash-resume |
| `pm_plan` | Reads brief → writes focused Agent 1 brief |
| `run_research` | Agent 1 — market, ICP, competitors (web search) |
| `run_brand` | Agent 2 — positioning, voice, pillars |
| `run_ux` | Agent 3 — flows, screens, wireframes, prototype prompts |
| `run_gtm` | Agent 4 — social (4a) + B2B outreach (4b) in parallel |
| `request_approval` | Human-in-the-loop checkpoint (stdin) |
| `assemble_plan` | Stitches every stage into `final_product_plan.md` |

## Install

### From source

```bash
git clone <this repo>
cd relay
go install .
```

The binary lands at `$(go env GOPATH)/bin/relay` — make sure that
directory is on your `$PATH`.

### Pre-built binaries

```bash
make release
# → dist/RELAY-darwin-arm64
# → dist/RELAY-darwin-amd64
# → dist/RELAY-linux-amd64
# → dist/RELAY-linux-arm64
# → dist/RELAY-windows-amd64.exe
```

On Windows without `make`, use the equivalent PowerShell loop in `Makefile`.

## Configure your MCP client

Set `ANTHROPIC_API_KEY` in the environment before launching the server.

### Claude Code — `.mcp.json`

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay",
      "env": { "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}" }
    }
  }
}
```

### VS Code + Copilot — `.vscode/mcp.json`

```json
{
  "servers": {
    "relay": {
      "command": "relay",
      "env": { "ANTHROPIC_API_KEY": "${env:ANTHROPIC_API_KEY}" }
    }
  }
}
```

### GitHub Copilot CLI — `~/.copilot/mcp-config.json`

```json
{
  "mcpServers": {
    "relay": {
      "type": "stdio",
      "command": "relay",
      "env": { "ANTHROPIC_API_KEY": "YOUR_KEY_HERE" }
    }
  }
}
```

### Cursor / Windsurf

Same shape as Claude Code's `.mcp.json` above.

## Usage

1. Drop a `product_brief.md` in the directory you want outputs written to.
2. From your MCP client, call `run_workflow` with `brief_path: "./product_brief.md"`.
3. The server logs progress to stderr and blocks at each H1–H4 checkpoint
   waiting for your decision on stdin.
4. When done you'll find `./output/final_product_plan.md`.

If the process crashes mid-pipeline, just call `run_workflow` again with the
same brief path — completed stages are skipped via `output/.session.meta.json`.

## Transports

The server speaks two transports. Default is **stdio** (best for local tools
like Claude Code, Cursor, Copilot CLI). Pass `--http` to expose the same 8
tools over Streamable-HTTP for remote clients (e.g. Claude.ai connectors).

```bash
# stdio (default) — what every MCP client config above expects
relay

# Streamable-HTTP on :8080 (default)
relay --http

# Custom bind address
relay --http --addr 0.0.0.0:9000
```

The HTTP endpoint lives at `/mcp`. SIGINT / SIGTERM trigger graceful shutdown
with a 5-second drain.

> **Note on HTTP mode:** stdin is unused, so checkpoint prompts auto-approve
> (the same behaviour as any non-TTY stdio session). This is the right
> default for an unattended remote server but means you do not get
> interactive iterate/skip control over HTTP — drive the pipeline via
> stdio if you need to iterate.

### Claude.ai connector

Expose the binary publicly (e.g. behind your reverse proxy of choice) and add
a connector pointing at `https://your-host/mcp`. The transport enforces an
MCP `initialize` handshake before any tool call, so the endpoint is safe to
expose — unauthenticated `tools/list` calls receive HTTP 404 ("Invalid
session ID").

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `ANTHROPIC_API_KEY` | *(required)* | Server exits non-zero on startup if missing |
| `MAX_ITERATIONS_PER_STAGE` | `5` | Hard cap on iterate loops; auto-advances with warning when reached |
| `CHECKPOINT_TIMEOUT_MINUTES` | `0` (off) | Auto-approve the checkpoint after N minutes of inactivity |

## Design guarantees

- **Atomic writes** — every output file goes through `WriteFile(.tmp) → Rename()`.
- **PID lockfiles** — concurrent calls to the same agent return an error
  rather than corrupting output. Stale locks (dead PID) are auto-cleared.
- **Context-window guard** — `ctxguard.Build` keeps total prompt under 120k
  chars; optional sections are dropped first, with a slog warning.
- **Crash-resume** — `output/.session.meta.json` tracks `completedStages`;
  re-running `run_workflow` skips them.
- **Non-interactive auto-approve** — when stdin is not a TTY, checkpoints
  auto-approve so the pipeline runs cleanly in CI.
- **Stdout cleanliness** — all logs go to stderr; stdout is reserved for the
  MCP JSON-RPC wire (verified by Phase 0's connection test).
- **Parallel GTM with panic recovery** — Agent 4a + 4b run concurrently;
  goroutine panics are recovered and surfaced as tool errors.

## Development

```bash
make build          # local binary
make test           # full test suite (no API key required)
make lint           # go vet + gofmt -l
```

The test suite runs without an `ANTHROPIC_API_KEY` — LLM-call paths are
exercised with a bogus key and asserted to surface as `IsError=true` results,
which is sufficient to verify wiring. Content quality requires a real key
and is verified by hand at the H1–H4 checkpoints.

## Layout

```
.
├── main.go                       MCP server entrypoint, registers 8 tools
├── tools/                        One file per tool + tests
├── internal/
│   ├── claude/                   Anthropic SDK wrapper (Call, CallWithSearch, CallJSON)
│   ├── ctxguard/                 120k-char window guard + multi-part Build
│   ├── logger/                   slog → stderr only
│   └── state/                    Atomic writes, PID locks, session meta
├── prompts/                      //go:embed *.md system prompts
├── Makefile                      build / test / release
└── dist/                         Cross-platform release binaries
```

## License

(Add a license file if distributing publicly.)

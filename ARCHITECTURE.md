# Architecture

## Overview

Relay is an MCP server that gives Claude a persistent workflow: product planning, research, brand voice, UX, and go-to-market -- all as structured tool calls. It runs as a stdio or HTTP MCP server, registers 40+ tools across 7 categories, and orchestrates multi-stage workflows with human checkpoints.

```
                    +-----------+
                    |  Claude   |
                    |  (client) |
                    +-----+-----+
                          |
                    MCP (stdio / http)
                          |
                    +-----v-----+
                    |  relay    |
                    |  server   |
                    +-----+-----+
                          |
          +---------------+---------------+----------------+
          |               |               |               |
    +-----v----+   +-------v-----+  +------v------+  +-----v-----+
    | registry |   |   claude    |  |   state     |  |  ctxguard  |
    | (tools)  |   |   client    |  |  (session)  |  | (context  |
    |          |   |  + retry    |  |             |  |  guard)   |
    +----+-----+   +------+------+  +-------------+  +-----------+
         |                |
    +----v----+    +------v------+
    |  tools  |    |   search    |
    | (7 cat) |    |  (DDG)      |
    +---------+    +-------------+
```

## Design Principles

1. **MCP-native, not a wrapper.** Relay is a real MCP server using `mcp-go`. Tools are registered with schemas, handlers return structured results. Claude sees them as first-class tools, not text commands.
2. **Human-in-the-loop by default.** Every workflow stage has a checkpoint. `request_approval` gates the pipeline until a human reviews and approves the output. No stage runs without explicit sign-off.
3. **Context is sacred.** The `ctxguard` package truncates tool inputs and outputs to stay within the model's context window. Required content survives; optional content is dropped first.
4. **Boring tech.** Go stdlib, `mcp-go`, SQLite-free (state is JSON on disk). No frameworks, no ORMs, no message queues.

## Components

### registry (`internal/registry`)

Thread-safe tool registry. Tools are registered at init time with a category, MCP tool definition, and handler function. The registry uses `sync.RWMutex` for concurrent read access.

- `Register(category, definition, handler)` -- adds a tool to the registry.
- `RegisterAll(server)` -- wires all registered tools into an `mcp-go` MCPServer.
- `List()` / `ListByCategory(cat)` -- returns tool definitions for display or filtering.
- `Count()` -- total registered tool count.

Categories: `file`, `image`, `pdf`, `data`, `text`, `web`, `workflow`.

### claude (`internal/claude`)

Anthropic API client with retry, streaming, and web search fallback.

**Retry strategy:** exponential backoff (2^attempt * 2s base delay, max 3 retries). Only retries on rate limits (429/529) and timeouts. Non-retryable errors fail fast.

**Streaming:** uses `Messages.NewStreaming` with 30-second progress logs. Accumulates events into a `Message` struct and extracts text blocks from the response.

**Web search (two-path):**
1. Native Anthropic web search tool (if not using a proxy).
2. DuckDuckGo fallback (if using a custom base URL that lacks web search support). Generates 3-5 search queries via a separate LLM call, runs `DDGMulti` to fetch results, and injects them as `<web_search_results>` in the user prompt with explicit untrusted-data warnings.

**JSON mode:** `CallJSON` wraps `Call` with a two-attempt JSON parser. Strips markdown fences, trims whitespace. On first parse failure, retries with a stricter prompt. On second failure, returns `ErrParseJSON`.

**Error classification:**
| HTTP pattern | Error kind | Retryable |
|---|---|---|
| 429, 529, "rate" | `rate_limit` | yes |
| `context.DeadlineExceeded` | `timeout` | yes |
| anything else | `api_error` | no |

### state (`internal/state`)

JSON-on-disk session state. Tracks workflow progress across stages.

**Session metadata** (`output/.session.meta.json`):
- `completedStages` -- ordered list of finished stages.
- `iterationCounts` -- per-stage iteration counter.
- `humanNotes` -- per-checkpoint notes from the human reviewer.

**File locking:** `AcquireLock` / `ReleaseLock` use PID-based lock files. Stale locks (process no longer running) are auto-cleaned.

**Atomic writes:** `WriteOutput` writes to a `.tmp` file first, then `os.Rename` for atomic replacement. Prevents partial reads if the process crashes mid-write.

**Stages:** `pm_plan`, `research`, `brand`, `ux`, `gtm`, `assembled`.

### ctxguard (`internal/ctxguard`)

Context window guard. Prevents tool inputs and outputs from exceeding the model's context limit.

- `maxChars = 120,000` -- hard limit per content block.
- `summaryChars = 3,000` -- how much to show when truncating.

`Guard(content, label)` -- returns content if under limit, otherwise returns the first 3,000 chars with a truncation notice.

`Build(parts)` -- assembles required and optional content blocks separated by `---`. If total exceeds `maxChars`, drops optional blocks and returns required only.

### search (`internal/search`)

DuckDuckGo HTML scraping (no API key needed).

- `DDG(ctx, query, limit)` -- POST to `html.duckduckgo.com/html/`, regex-parse results.
- `DDGMulti(ctx, queries, perQuery)` -- runs multiple queries, deduplicates by URL.
- `FormatForPrompt(results)` -- wraps results in `<result>` XML tags with index, title, URL, snippet.
- `sanitizeForPrompt` -- strips `--- WEB_SEARCH_RESULTS ---` markers and `<result>` tags from snippets to prevent prompt injection via search content.

### license (`internal/license`)

Ed25519-signed license verification. Closed-beta gating.

- License format: `RELAY-<payload>.<signature>` where payload is base64url-encoded JSON with `sub`, `iat`, `exp` fields.
- Public key is embedded in the binary via `//go:embed`.
- Verification: parse PEM, `ed25519.Verify`, check expiry (24-hour grace period).
- Sources: `$RELAY_LICENSE` env var, or `~/.relay/license` file.
- `Sign` function (in `cmd/genkeys`) generates new licenses with a private key.

### logger (`internal/logger`)

Thin wrapper around `log/slog` with structured logging. Output goes to stderr (so it does not interfere with stdio MCP transport). `Quiet` flag suppresses info-level during normal operation.

## Tool Categories

| Category | Tools | Purpose |
|---|---|---|
| file | `file_hash`, `file_read`, `file_write`, `file_list`, `file_size`, `file_zip` | Filesystem operations |
| image | `image_resize`, `image_convert`, `image_metadata` | Image processing |
| pdf | `pdf_extract`, `pdf_merge` | PDF operations |
| data | `data_csv_parse`, `data_json_query`, `data_json_validate` | Data transformation |
| text | `text_count`, `text_diff`, `text_template` | Text processing |
| web | `web_fetch`, `web_search` | Web access (DDG + HTTP fetch) |
| workflow | `run_pm_plan`, `run_research`, `run_brand`, `run_ux`, `run_gtm`, `assemble_plan`, `request_approval` | Multi-stage pipeline |

## Workflow Pipeline

```
product_brief.md
       |
       v
  +----------+     +-----------+     +--------+     +-----+     +-----+
  | pm_plan  | --> | research  | --> | brand  | --> | ux  | --> | gtm |
  +----------+     +-----------+     +--------+     +-----+     +-----+
       |              |                |             |           |
       v              v                v             v           v
   output/         output/         output/       output/     output/
   pm_plan.md      research.md     brand.md      ux.md       gtm.md
       |                                                          |
       +---------------------> assemble_plan <---------------------+
                                   |
                                   v
                            output/assembled.md
```

Each stage:
1. Reads `product_brief.md` + prior stage outputs.
2. Calls `claude.CallWithSearch` (LLM + web search).
3. Writes output to `output/<stage>.md` via `state.WriteOutput`.
4. Calls `request_approval` -- human reviews, approves, or requests iteration.
5. On approval, `state.MarkStageComplete` records the stage.

## Process Model

Relay runs as a single process. Two transport modes:
- **stdio** (default): MCP over stdin/stdout. Claude Desktop or CLI connects directly.
- **HTTP** (`--http`): MCP over HTTP. Listens on `$RELAY_HTTP_ADDR` (default `:8080`).

No daemon. No background service. The process starts, registers tools, and serves until the client disconnects or a signal is received.

## File Layout

```
project/
  product_brief.md          # Input: project description
  .env                      # Optional env vars
  output/
    .session.meta.json       # Session state
    pm_plan.md               # Stage outputs
    research.md
    brand.md
    ux.md
    gtm.md
    assembled.md
    *.lock                   # Per-file PID locks
```

## Testing

60 tests, 67% coverage. Race detector enabled in CI (`-race` flag). Tests cover:
- ctxguard: 22 tests (truncation, build, edge cases)
- registry: thread safety, category filtering
- claude: error classification, JSON parsing, retry logic
- state: session lifecycle, lock management, atomic writes
- search: DDG parsing, URL cleaning, deduplication

## Dependencies

- `github.com/mark3labs/mcp-go` -- MCP server implementation
- `github.com/anthropics/anthropic-sdk-go` -- Anthropic API client
- `github.com/joho/godotenv` -- .env file loading
- Go stdlib for everything else

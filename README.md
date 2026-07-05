<div align="center">

# relay

**Give your AI agent one local MCP server that can actually do useful work.**

[![CI](https://img.shields.io/github/actions/workflow/status/valtors/relay/ci.yml?style=flat&label=ci)](https://github.com/valtors/relay/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/valtors/relay)](go.mod)
[![Contributors welcome](https://img.shields.io/badge/contributors-welcome-brightgreen.svg)](CONTRIBUTING.md)

[Install](#install) · [5-minute quickstart](#5-minute-quickstart) · [Workflows](#3-workflows-to-try-first) · [Client configs](#client-configs) · [Contributing](CONTRIBUTING.md)

</div>

Relay is a local-first MCP server for Claude Desktop, Cursor, VS Code, and other MCP clients.

Install one Go binary and your agent can:

- read and write local files
- resize, crop, and convert images
- extract, merge, split, and inspect PDFs
- fetch web pages and check links
- convert CSV, JSON, markdown, base64, and regex output

No pile of single-purpose servers. No plugin hunt. Just one binary with 40 built-in tools.

## Why people use Relay

- **Fast first run** - install it, run `relay init`, restart your editor, start using it
- **Local-first** - your files stay on your machine
- **Broad utility** - file, image, PDF, text, data, and web tools in one place
- **Simple ops** - single Go binary, cross-platform releases, no extra runtime to manage

## 5-minute quickstart

### 1) Install

### Fastest way (recommended)

```bash
npx userelay
```

That's it. Downloads the binary, starts the server. Works on macOS, Linux, and Windows.

To set up your editor:

```bash
npx userelay init
```

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/valtors/relay/main/scripts/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/valtors/relay/main/scripts/install.ps1 | iex
```

**From source**

```bash
go install github.com/valtors/relay@latest
```

### 2) Verify the binary

```bash
relay status
```

Expected shape:

```text
relay v<version>
tools: 40 registered (7 categories)
transport: stdio (default) | http (with --http)
```

### 3) Connect Relay to your editor

Fastest path:

```bash
relay init
```

Relay detects supported editors, writes the config entry, and tells you what to restart.

### 4) Ask for one useful result

Try this in your MCP client:

> Resize `./screenshots/hero.png` to 1200px wide and save it as `./screenshots/hero-large.png`.

If that works, Relay is live.

---

## 3 workflows to try first

These are the workflows that should convert a new user fast.

### 1) Research a repo or docs page

**Prompt**

> Fetch `https://github.com/pdfcpu/pdfcpu`, summarize what it does, list 3 commands worth trying, and save the notes to `./notes/pdfcpu-summary.md`.

**What Relay helps with**

- `web_fetch`
- `file_write`
- text formatting tools as needed

**Expected result**

A local markdown summary your agent can reuse later.

### 2) Turn local PDFs into structured data

**Prompt**

> Read every PDF in `./invoices`, extract the text, pull out all dollar amounts, and save one JSON file at `./invoices/amounts.json`.

**What Relay helps with**

- `file_list`
- `pdf_extract_text`
- `data_json_format`
- `file_write`

**Expected result**

A clean JSON artifact instead of manual copy-paste from PDFs.

### 3) Fetch the web and ship a usable artifact

**Prompt**

> Fetch `https://example.com/pricing`, turn the important points into markdown, convert that markdown to HTML, and save it as `./research/pricing.html`.

**What Relay helps with**

- `web_fetch`
- `text_md_to_html`
- `file_write`

**Expected result**

A file you can open, review, and share immediately.

---

## Client configs

**Best option:** use `relay init`.

That is the most reliable path because it writes the right JSON shape for the detected editor and uses the actual installed binary path.

If you want to paste config manually, use the snippets below. They assume `relay` is already on your `PATH`. If your client cannot find it, replace `"relay"` with the full path from `which relay` or `where relay`.

### Claude Desktop

**macOS**
`~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows**
`%APPDATA%\Claude\claude_desktop_config.json`

**Linux**
`~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay"
    }
  }
}
```

### Cursor

**Project config**
`<project>/.cursor/mcp.json`

**Global config**
`~/.cursor/mcp.json` or `%USERPROFILE%\.cursor\mcp.json`

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay"
    }
  }
}
```

### VS Code

**Project config**
`<project>/.vscode/mcp.json`

```json
{
  "servers": {
    "relay": {
      "command": "relay"
    }
  }
}
```

### Optional: enable Relay workflow tools

Most Relay tools do **not** need an API key.

If you want the built-in planning and workflow tools (`run_workflow`, `pm_plan`, `run_research`, `run_brand`, `run_ux`, `run_gtm`, `assemble_plan`), add `ANTHROPIC_API_KEY` under `env` in your client config:

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay",
      "env": {
        "ANTHROPIC_API_KEY": "your-key-here"
      }
    }
  }
}
```

---

## What Relay does today

Relay ships with **40 tools across 7 categories**.

### Local utility tools

- **File (7)** - read, write, list, size, hash, zip, unzip
- **Image (7)** - info, resize, crop, convert, rotate, grayscale, flip
- **PDF (6)** - info, extract text, page count, merge, split, extract pages
- **Text (6)** - word count, replace, regex extract, base64 encode/decode, markdown to HTML
- **Data (4)** - format JSON, CSV to JSON, JSON to CSV, JSON query
- **Web (2)** - fetch page content, check status

### Workflow tools

- **Workflow (8)** - higher-level planning and orchestration helpers for product/strategy flows

See everything:

```bash
relay tools
relay tools --json
```

---

## Why Relay instead of stitching servers together?

| | Relay | Typical setup |
|---|---|---|
| Install | One binary | Multiple repos, runtimes, and configs |
| Scope | 40 built-in tools | Usually one narrow tool per server |
| Setup time | Minutes | Often a half hour of glue work |
| Local file workflows | First-class | Varies |
| Cross-platform release | Yes | Inconsistent |
| Ops overhead | Low | Higher |

Relay is the practical option if you want one MCP server that covers the common local tasks an agent actually needs.

---

## CLI

```bash
relay              # start MCP server in stdio mode
relay start        # same as above
relay start --http # serve over Streamable HTTP
relay init         # detect editor and write config
relay init --list  # show detected editors
relay tools        # list all tools
relay tools --json # list all tools as JSON
relay status       # version, tool count, transport
relay version      # print version
relay help         # usage info
```

---

## Development

```bash
git clone https://github.com/valtors/relay
cd relay
go test ./...
go run . status
go run .
```

If you want to add a tool, start here:

- [docs/ADDING_A_TOOL.md](docs/ADDING_A_TOOL.md)

---

## Roadmap

What is solid now:

- 40 built-in tools
- cross-platform install scripts
- MCP server over stdio and HTTP
- editor config helper with `relay init`
- GoReleaser and CI

What is next:

- tighter onboarding assets
- demo GIF
- more verified starter workflows
- deeper orchestration where it clearly improves real use

---

## License

MIT. See [LICENSE](LICENSE).

---

<div align="center">

Built by [Tamish](https://github.com/tamish560) at [Valtors](https://github.com/valtors)

If Relay saves you setup time, give it a star.

</div>

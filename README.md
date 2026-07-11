<div align="center">

<img src="https://raw.githubusercontent.com/valtors/relay-landing/master/assets/images/relay-logo.png" alt="Relay logo" width="160" />

# Relay

Local MCP server for files, images, PDFs, web, and workflows.

[![CI](https://img.shields.io/github/actions/workflow/status/valtors/relay/ci.yml?style=for-the-badge&logo=githubactions&label=ci)](https://github.com/valtors/relay/actions/workflows/ci.yml)
[![npm version](https://img.shields.io/npm/v/userelay?style=for-the-badge&logo=npm&label=npm)](https://www.npmjs.com/package/userelay)
[![License](https://img.shields.io/badge/license-MIT-16a34a?style=for-the-badge)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/valtors/relay?style=for-the-badge&logo=go)](go.mod)
[![GitHub stars](https://img.shields.io/github/stars/valtors/relay?style=for-the-badge&logo=github)](https://github.com/valtors/relay/stargazers)
[![npm downloads](https://img.shields.io/npm/dm/userelay?style=for-the-badge&logo=npm&label=downloads)](https://www.npmjs.com/package/userelay)
[![GitHub release](https://img.shields.io/github/v/release/valtors/relay?style=for-the-badge&logo=github&label=release)](https://github.com/valtors/relay/releases)
[![good first issue](https://img.shields.io/github/issues/valtors/relay/good%20first%20issue?style=for-the-badge&label=good%20first%20issue&color=7057ff)](https://github.com/valtors/relay/labels/good%20first%20issue)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-7057ff?style=for-the-badge)](CONTRIBUTING.md)

</div>

---

## New here?

- **First time?** Run `npx userelay` and follow the animated setup wizard. No install needed. It will download Relay and configure your editor.
- **Already set up?** Run `npx userelay tui` for the interactive menu.
- **Want to reconfigure your editor?** Run `npx userelay init`.
- **Just curious?** Run `npx userelay --help` or `npx userelay status`.
- **Want to contribute?** Read [Your first PR in 5 minutes](docs/FIRST_PR.md), check [`good first issues`](https://github.com/valtors/relay/labels/good%20first%20issue), or see [how to add a tool](docs/ADDING_A_TOOL.md).
- **Want to ask first?** Open a [Discussion](https://github.com/valtors/relay/discussions). We reply within 24 hours.

---

## Why Relay?

Every MCP server does one thing. File servers do files. Image servers do images. PDF servers do PDFs. Before Relay, running a capable agent meant stitching together 5-10 separate servers, each with its own install, config, and update cycle.

Relay puts 40 tools behind one binary. One install. One config. One process.

| | Relay | Multiple servers |
|---|---|---|
| Install | `npx userelay` | One per server |
| Config | One entry | One per server |
| Updates | One binary | One per server |
| Memory | One process | N processes |
| Tools | 40 built-in | Whatever you assembled |

---

## Install

```bash
npx userelay tui
```

Runs locally. Your files stay on your machine. Type `q` to exit.

---

## What it does

- Runs one MCP server with 40 tools across 7 categories.
- Reads and writes local files.
- Resizes, crops, converts, rotates, grayscales, and flips images.
- Extracts text from PDFs, counts pages, merges files, splits files, and extracts pages.
- Fetches web pages and checks status codes.
- Converts CSV, JSON, markdown, base64, and regex output.
- Works with Claude Desktop, Cursor, VS Code, and other MCP clients.
- Supports stdio by default and HTTP with `--http`.
- Enables built-in workflow tools when `ANTHROPIC_API_KEY` is present.

No pile of single-purpose servers. One binary.

---

## Quickstart

### 1. Check the binary

Use `status` first if you want a one-shot command instead of starting the stdio server.

```bash
npx userelay status
```

```text
relay v0.3.0
tools: 40 registered (7 categories)
transport: stdio (default) | http (with --http)
```

### 2. Detect supported editors

```bash
npx userelay init --list
```

```text
relay init --list

detected editors:
  1. Cursor
  2. VS Code
```

If your editor is listed, write the config:

```bash
npx userelay init
```

### 3. Verify the local binary

If you installed Relay from source or a release, `relay status` prints the same shape:

```bash
relay status
```

```text
  ╭─────────────────────────────────╮
  │  relay vdev                     │
  │                                 │
  │  Tools:      40 (7 categories)  │
  │  Transport:  stdio | http       │
  │  Status:     ready              │
  ╰─────────────────────────────────╯
```

### 4. Try one result

Prompt your MCP client with this:

> Resize `./screenshots/hero.png` to 1200px wide and save it as `./screenshots/hero-large.png`.

If that works, Relay is live.

---

## Workflows to try first

### 1. Research a repo or docs page

**Prompt**

> Fetch `https://github.com/pdfcpu/pdfcpu`, summarize what it does, list 3 commands worth trying, and save the notes to `./notes/pdfcpu-summary.md`.

**Relay tools involved**

- `web_fetch`
- `file_write`
- text formatting tools as needed

**Expected result**

A local markdown summary your agent can reuse later.

### 2. Turn local PDFs into structured data

**Prompt**

> Read every PDF in `./invoices`, extract the text, pull out all dollar amounts, and save one JSON file at `./invoices/amounts.json`.

**Relay tools involved**

- `file_list`
- `pdf_extract_text`
- `data_json_format`
- `file_write`

**Expected result**

A JSON artifact instead of manual copy-paste from PDFs.

### 3. Fetch the web and ship a usable artifact

**Prompt**

> Fetch `https://example.com/pricing`, turn the important points into markdown, convert that markdown to HTML, and save it as `./research/pricing.html`.

**Relay tools involved**

- `web_fetch`
- `text_md_to_html`
- `file_write`

**Expected result**

A file you can open, review, and share.

---

## Tools

Relay ships with 40 tools across 7 categories.

| Category | Tools |
|---|---|
| File (7) | read, write, list, size, hash, zip, unzip |
| Image (7) | info, resize, crop, convert, rotate, grayscale, flip |
| PDF (6) | info, page count, extract text, extract pages, merge, split |
| Text (6) | word count, replace, extract regex, base64 encode, base64 decode, md to html |
| Data (4) | csv to json, json to csv, format json, query json |
| Web (2) | fetch, status |
| Workflow (8) | run workflow, PM plan, research, brand, UX, GTM, approval, assemble plan |

List them from the CLI:

```bash
relay tools
relay tools --json
```

---

## Client configs

Best option: use `relay init`.

That writes the right JSON shape for the detected editor and uses the installed binary path. If you want to paste config manually, use the snippets below. They assume `relay` is already on your `PATH`. If your client cannot find it, replace `"relay"` with the full path from `which relay` or `where relay`.

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

### Optional workflow tools

Most Relay tools do not need an API key.

If you want the built-in planning and workflow tools (`run_workflow`, `pm_plan`, `run_research`, `run_brand`, `run_ux`, `run_gtm`, `assemble_plan`), add `ANTHROPIC_API_KEY` under `env`:

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

## Comparison

| Capability | Relay | Stitching servers together |
|---|---:|---:|
| One-command install | ✓ | — |
| One repo / one binary | ✓ | — |
| 40 built-in tools | ✓ | — |
| Minutes to first result | ✓ | — |
| Local file workflows | ✓ | — |
| Cross-platform release | ✓ | — |
| Low ops overhead | ✓ | — |

---

## CLI reference

```text
relay              start MCP server in stdio mode
relay tui          launch interactive TUI (same as no args in a TTY)
relay start        same as above
relay start --http serve over Streamable HTTP
relay init         detect editor and write config
relay init --list  show detected editors
relay tools        list all tools
relay tools --json list all tools as JSON
relay status       version, tool count, transport
relay version      print version
relay help         usage info
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

Add a tool: [docs/ADDING_A_TOOL.md](docs/ADDING_A_TOOL.md)

Install alternatives:

```bash
curl -fsSL https://raw.githubusercontent.com/valtors/relay/main/scripts/install.sh | sh
```

```powershell
irm https://raw.githubusercontent.com/valtors/relay/main/scripts/install.ps1 | iex
```

```bash
go install github.com/valtors/relay@latest
```

---

## Interactive mode

When you run `npx userelay` without arguments in a terminal, Relay launches an interactive TUI:

```text
  ╭───────────────────────────────────╮
  │  relay v0.3.0                    │
  │  40 tools · 7 categories         │
  ╰───────────────────────────────────╯

  ❯ Start MCP server (stdio)
    Start MCP server (HTTP)
    Initialize — detect & configure editors
    Browse tools
    Status
    Quit
```

- Animated RELAY wordmark on launch (shimmer effect)
- Keyboard-navigable menus
- Init wizard with editor detection and config writing
- Tools browser — all 40 tools organized by category
- Status dashboard — version, transport, tool count

If stdin is not a TTY (piped, CI, or MCP client), Relay starts the server directly — no TUI.

To explicitly launch the TUI: `npx userelay tui`

---

## Roadmap

**Shipped:**
- 40 built-in tools across 7 categories
- Cross-platform install (npm, curl, go install)
- MCP server over stdio and HTTP
- Editor config helper (`relay init`)
- Interactive TUI mode
- Security hardening (path traversal, SSRF, XSS, prompt injection)
- GoReleaser CI/CD

**Next:**
- Tool plugin system (define tools in external files)
- Streaming responses for long-running operations
- More verified starter workflows
- Community-contributed tools
- MCP server discovery and registry

**Thinking about:**
- Docker image for containerized deployments
- WebSocket transport
- Tool-level permissions and sandboxing
- Multi-language tool definitions (Python, JS plugins)

---

## Contributors

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/tamish560"><img src="https://avatars.githubusercontent.com/u/189916421?v=4" width="80px;" alt=""/><br /><sub><b>Tamish Mhatre</b></sub></a><br /><a title="Code">💻</a> <a title="Design">🎨</a> <a title="Docs">📖</a></td>
  </tr>
</table>
<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

Want to be here? Check [`good first issues`](https://github.com/valtors/relay/labels/good%20first%20issue).

---

## License

MIT. See [LICENSE](LICENSE).

Built by [Tamish](https://github.com/tamish560) at [Valtors](https://github.com/valtors). If Relay saves setup time, star the repo.

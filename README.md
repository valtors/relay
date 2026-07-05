<div align="center">

<img src="https://raw.githubusercontent.com/valtors/relay-landing/master/assets/images/relay-logo.png" alt="Relay mascot" width="160" />

# Relay

### One local-first MCP server for files, images, PDFs, web, and workflow magic ✨

```text
┌─────────────────────────────────────┐
│  npx userelay                       │
│  That's it. You're ready.           │
└─────────────────────────────────────┘
```

[![npm version](https://img.shields.io/npm/v/userelay?style=for-the-badge&logo=npm&label=npm)](https://www.npmjs.com/package/userelay)
[![npm downloads](https://img.shields.io/npm/dm/userelay?style=for-the-badge&logo=npm&label=downloads)](https://www.npmjs.com/package/userelay)
[![GitHub stars](https://img.shields.io/github/stars/valtors/relay?style=for-the-badge&logo=github)](https://github.com/valtors/relay/stargazers)
[![GitHub release](https://img.shields.io/github/v/release/valtors/relay?style=for-the-badge&logo=github&label=release)](https://github.com/valtors/relay/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/valtors/relay/ci.yml?style=for-the-badge&logo=githubactions&label=ci)](https://github.com/valtors/relay/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-16a34a?style=for-the-badge)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/valtors/relay?style=for-the-badge&logo=go)](go.mod)
[![Contributors welcome](https://img.shields.io/badge/contributors-welcome-brightgreen?style=for-the-badge)](CONTRIBUTING.md)

**[Quickstart](#quickstart) · [Tools](#tools) · [Why Relay](#why-relay) · [Client Configs](#client-configs) · [Roadmap](#roadmap)**

</div>

Relay is a local-first MCP server for Claude Desktop, Cursor, VS Code, and other MCP clients.

Install one Go binary and your agent can:

- read and write local files
- resize, crop, and convert images
- extract, merge, split, and inspect PDFs
- fetch web pages and check links
- convert CSV, JSON, markdown, base64, and regex output

No pile of single-purpose servers. No plugin hunt. Just one binary with 40 built-in tools.

<a id="why-relay"></a>

## ⚡ Why Relay

- **Fast first run** - install it, run `relay init`, restart your editor, start using it
- **Local-first** - your files stay on your machine
- **Broad utility** - file, image, PDF, text, data, and web tools in one place
- **Simple ops** - single Go binary, cross-platform releases, no extra runtime to manage

### Relay vs stitching servers together

| Capability | Relay | Stitching servers together |
|---|---:|---:|
| One-command install | ✅ | ❌ |
| One repo / one binary | ✅ | ❌ |
| 40 built-in tools | ✅ | ❌ |
| Minutes to first result | ✅ | ❌ |
| Local file workflows | ✅ | ❌ |
| Cross-platform release | ✅ | ❌ |
| Low ops overhead | ✅ | ❌ |

Relay is the practical option if you want one MCP server that covers the common local tasks an agent actually needs.

<a id="quickstart"></a>

## 🚀 Quickstart

### The hero moment

```bash
npx userelay
```

That is the fast path. Relay downloads the binary, starts the server, and gets you from zero to useful in one command.

### 1) Install

#### Fastest way (recommended)

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

<a id="client-configs"></a>

## 📦 Client Configs

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

<a id="tools"></a>

## 🛠️ Tools

Relay ships with **40 tools across 7 categories**.

| Category | Included tools |
|---|---|
| 📁 **File (7)** | read, write, list, size, hash, zip, unzip |
| 🖼️ **Image (7)** | info, resize, crop, convert, rotate, grayscale, flip |
| 📄 **PDF (6)** | info, extract, pages, merge, split, extract pages |
| ✏️ **Text (6)** | word count, replace, regex, base64, md→html |
| 📊 **Data (4)** | JSON format, CSV↔JSON, query |
| 🌐 **Web (2)** | fetch, status check |
| 🔄 **Workflow (8)** | planning & orchestration |

See everything:

```bash
relay tools
relay tools --json
```

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

<a id="roadmap"></a>

## 🗺️ Roadmap

### What is solid now

- 40 built-in tools
- cross-platform install scripts
- MCP server over stdio and HTTP
- editor config helper with `relay init`
- GoReleaser and CI

### What is next

- tighter onboarding assets
- demo GIF
- more verified starter workflows
- deeper orchestration where it clearly improves real use

---

## License

MIT. See [LICENSE](LICENSE).

---

<div align="center">

**Built with ❤️ by [Tamish](https://github.com/tamish560) at [Valtors](https://github.com/valtors)**  
If Relay saves you setup time, **give it a star** ⭐

</div>

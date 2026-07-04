<div align="center">

# relay

**it just figures out and gets it done.**

[![CI](https://img.shields.io/github/actions/workflow/status/valtors/relay/ci.yml?style=flat&label=ci)](https://github.com/valtors/relay/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/valtors/relay)](go.mod)
[![Contributors welcome](https://img.shields.io/badge/contributors-welcome-brightgreen.svg)](CONTRIBUTING.md)

One MCP server. 40 built-in tools. Single Go binary. Zero config.

Your AI agent gets file ops, image processing, PDF manipulation, web fetching, data conversion, and text transforms - all without you wiring up 12 different servers.

[Install in 30 seconds](#install) · [Try it now](#try-it-now) · [All 40 tools](#tools) · [Contributing](CONTRIBUTING.md)

</div>

---

## What is this?

Relay is an MCP server you plug into Claude Desktop, Cursor, VS Code Copilot, or any MCP client. Once connected, your AI agent can:

- Resize and crop images without Photoshop
- Extract text from PDFs and merge them
- Fetch web pages and check link status
- Hash files, zip folders, read and write anything
- Convert CSV to JSON, format data, query nested objects
- Encode base64, run regex, convert markdown to HTML

One binary. Runs locally. Your files never leave your machine.

---

## Install

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/valtors/relay/main/scripts/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/valtors/relay/main/scripts/install.ps1 | iex
```

**From source:**
```bash
go install github.com/valtors/relay@latest
```

---

## Connect to your editor

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-your-key-here"
      }
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json` in your project:

```json
{
  "mcpServers": {
    "relay": {
      "command": "relay",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-your-key-here"
      }
    }
  }
}
```

### VS Code (Copilot)

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "relay": {
      "command": "relay",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-your-key-here"
      }
    }
  }
}
```

Restart your editor. Relay is now available.

---

## Try it now

Once connected, just ask your agent naturally. Relay figures out which tool to use.

### Workflow 1: Research a codebase

> "Fetch the README from https://github.com/pdfcpu/pdfcpu and summarize what it does"

Relay uses `web_fetch` to grab the page, your agent reads it, done.

### Workflow 2: Process local files

> "Take all the PNGs in ./screenshots, resize them to 800px wide, and put them in ./resized"

Relay uses `file_list` to find the PNGs, `image_resize` for each one, writes them out.

### Workflow 3: Extract and convert data

> "Read invoice.pdf, extract the text, then pull out all dollar amounts into a JSON array"

Relay uses `pdf_extract_text` to get the content, your agent parses it, `data_json_format` cleans up the output.

---

## Tools

40 tools across 7 categories. All built-in, no plugins needed.

### File (7)
| Tool | What it does |
|------|-------------|
| `file_read` | Read file contents |
| `file_write` | Write content to a file |
| `file_list` | List directory contents |
| `file_size` | Get file size |
| `file_hash` | SHA-256 hash of a file |
| `file_zip` | Create a zip archive |
| `file_unzip` | Extract a zip archive |

### Image (7)
| Tool | What it does |
|------|-------------|
| `image_info` | Get dimensions, format, file size |
| `image_resize` | Resize with proportional scaling |
| `image_crop` | Crop to rectangle |
| `image_convert` | Convert between PNG, JPEG, GIF |
| `image_rotate` | Rotate 90, 180, or 270 degrees |
| `image_grayscale` | Convert to grayscale |
| `image_flip` | Flip horizontal or vertical |

### PDF (6)
| Tool | What it does |
|------|-------------|
| `pdf_info` | Page count, title, author, dimensions |
| `pdf_extract_text` | Extract text from pages |
| `pdf_page_count` | Get number of pages |
| `pdf_merge` | Merge multiple PDFs |
| `pdf_split` | Split into individual pages |
| `pdf_extract_pages` | Extract specific pages |

### Text (6)
| Tool | What it does |
|------|-------------|
| `text_word_count` | Count words |
| `text_replace` | Find and replace |
| `text_extract_regex` | Extract regex matches |
| `text_base64_encode` | Base64 encode |
| `text_base64_decode` | Base64 decode |
| `text_md_to_html` | Markdown to HTML |

### Data (4)
| Tool | What it does |
|------|-------------|
| `data_json_format` | Pretty-print JSON |
| `data_csv_to_json` | CSV to JSON array |
| `data_json_to_csv` | JSON array to CSV |
| `data_json_query` | Query JSON by dot-path |

### Web (2)
| Tool | What it does |
|------|-------------|
| `web_fetch` | Fetch URL body |
| `web_status` | Check if URL is reachable |

### Workflow (8)
| Tool | What it does |
|------|-------------|
| `run_workflow` | Full multi-agent pipeline |
| `pm_plan` | Generate product brief |
| `run_research` | Market research |
| `run_brand` | Brand positioning |
| `run_ux` | UX wireframes |
| `run_gtm` | Go-to-market plan |
| `request_approval` | Human checkpoint |
| `assemble_plan` | Final plan assembly |

Run `relay tools` in your terminal to see all tools with descriptions.

---

## CLI

```bash
relay              # start MCP server (stdio mode)
relay start        # same as above
relay start --http # HTTP mode for web clients
relay tools        # list all tools
relay tools --json # tools as JSON
relay status       # version, tool count, categories
relay version      # print version
relay help         # usage info
```

---

## Why Relay over other MCP servers?

| | Relay | Others |
|---|---|---|
| Language | Go (single binary, fast, portable) | Usually Python or Node |
| Setup | One command | pip install + config + deps |
| Built-in tools | 40 ready to use | Bring your own |
| Workflow def | None needed, just talk | YAML roles, DAG graphs, code |
| MCP native | Yes, built for it | Usually added as plugin |
| Runs locally | Yes, your files stay private | Varies |
| Price | Free forever, MIT license | Free to paid |

---

## How it works under the hood

```
Your Editor (Claude / Cursor / VS Code)
    |
    | MCP protocol (stdio or HTTP)
    v
relay binary (~15 MB)
    |
    |-- tool registry (self-registering, grouped by category)
    |-- file/image/pdf/text/data/web handlers
    |-- workflow engine (multi-agent orchestration)
    |-- state persistence (JSON on disk)
    |
    v
Your local filesystem (nothing leaves your machine)
```

---

## Development

```bash
git clone https://github.com/valtors/relay
cd relay
go test ./...   # run all tests
go run .        # start locally
```

Adding a tool is the easiest contribution. See [docs/ADDING_A_TOOL.md](docs/ADDING_A_TOOL.md).

---

## Roadmap

- [x] 40 built-in tools (file, image, pdf, text, data, web)
- [x] CLI with subcommands
- [x] Self-registering tool system
- [x] Cross-platform binaries (GoReleaser)
- [x] Comprehensive test suite
- [ ] Orchestration (agent memory, checkpoints, handoffs)
- [ ] Hermes bridge (WhatsApp, Discord, Telegram)
- [ ] Plugin system for community tools
- [ ] Multi-LLM support

---

## License

MIT. See [LICENSE](LICENSE).

---

<div align="center">

Built by [Tamish](https://github.com/tamish560) at [Valtors](https://github.com/valtors)

**[Star on GitHub](https://github.com/valtors/relay)** if this is useful to you.

</div>

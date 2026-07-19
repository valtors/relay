# Migrating from standalone MCP servers to Relay

If you're currently running several separate MCP servers — one for files, one for
images, one for PDFs, one for fetching web pages — this guide shows how to
replace them all with a single Relay binary.

## Why consolidate?

| | Relay | Multiple standalone servers |
|---|---|---|
| Install | `npx userelay` (one command) | One install per server |
| Config | One entry in `claude_desktop_config.json` | One entry per server |
| Updates | One binary to update | One update per server |
| Memory / processes | One process | N processes running at once |
| Tools available | 40 built-in tools across 7 categories | Whatever you've individually assembled |

Every standalone server you're running today adds its own install step, its own
config block, and its own background process. Relay collapses all of that into
one binary with one config entry.

## Before: multiple servers in `claude_desktop_config.json`

A typical setup running several standalone MCP servers might look like this,
using the official fetch reference server from
[modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)
as an example:

```json
{
  "mcpServers": {
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
    // ...plus a separate entry for every other server you use
  }
}
```

Each entry is its own process, its own install, and its own thing to keep updated.

## After: one Relay entry

```json
{
  "mcpServers": {
    "relay": {
      "command": "npx",
      "args": ["-y", "userelay"]
    }
  }
}
```

Run `relay init` and it will detect your editor (Claude Desktop, Cursor, VS Code)
and write this for you automatically.

## Tool mapping by category

For editor-specific configs beyond Claude Desktop (Cursor, VS Code), or for
running Relay via Docker or as an HTTP server, see the
[`examples/`](../examples) directory in this repo.

Relay ships 40 tools across 7 categories. Here's what's available to replace
common standalone-server functionality:

### File operations

| Tool | What it does |
|---|---|
| `file_read` | Read file contents |
| `file_write` | Write content to a file |
| `file_list` | List directory contents (optionally recursive) |
| `file_size` | Get file size |
| `file_hash` | Compute a file's SHA-256 hash |
| `file_zip` | Create a zip archive |
| `file_unzip` | Extract a zip archive |

### Image operations

| Tool | What it does |
|---|---|
| `image_info` | Get image metadata (dimensions, format, size) |
| `image_resize` | Resize an image |
| `image_crop` | Crop an image to a rectangle |
| `image_convert` | Convert between png, jpeg, gif |
| `image_rotate` | Rotate 90/180/270 degrees |
| `image_grayscale` | Convert to grayscale |
| `image_flip` | Flip horizontally or vertically |

### PDF operations

| Tool | What it does |
|---|---|
| `pdf_info` | Get PDF metadata (page count, title, author, creator, page dimensions) |
| `pdf_extract_text` | Extract text from PDF pages |
| `pdf_page_count` | Get page count |
| `pdf_merge` | Merge multiple PDFs |
| `pdf_split` | Split a PDF into individual pages |
| `pdf_extract_pages` | Extract specific pages into a new PDF |

### Data operations

| Tool | What it does |
|---|---|
| `data_json_format` | Pretty-print JSON |
| `data_csv_to_json` | Convert CSV to JSON |
| `data_json_to_csv` | Convert JSON to CSV |
| `data_json_query` | Query JSON by dot path |

### Text operations

| Tool | What it does |
|---|---|
| `text_word_count` | Count words |
| `text_replace` | Find and replace |
| `text_extract_regex` | Extract regex matches |
| `text_base64_encode` / `text_base64_decode` | Base64 encode/decode |
| `text_md_to_html` | Convert markdown to HTML |

### Web operations

| Tool | What it does |
|---|---|
| `web_fetch` | Fetch a URL's body |
| `web_status` | Check URL reachability |

If you're currently running a standalone fetch-only MCP server (for example,
`mcp-server-fetch`), `web_fetch` replaces that functionality directly:

```json
// Before
{
  "mcpServers": {
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
```

```json
// After
{
  "mcpServers": {
    "relay": {
      "command": "npx",
      "args": ["-y", "userelay"]
    }
  }
}
```

### Workflow operations

Relay also includes higher-level workflow tools not typically found in
standalone single-purpose servers — these orchestrate multi-step processes
like product planning, research, and go-to-market work, gated behind an
`ANTHROPIC_API_KEY`. These aren't a 1:1 replacement for any standalone server,
so they're mentioned here for completeness rather than as a migration target.

## Summary

Replacing N standalone MCP servers with Relay means:
- One install (`npx userelay`) instead of N
- One config entry instead of N
- One process running instead of N
- Access to 40 tools instead of whatever subset you'd individually assembled

---
*Tool list verified against `tools/registrations.go` and the live binary (relay v0.4.10). Confirmed with the maintainer in [issue #30](https://github.com/valtors/relay/issues/30).*

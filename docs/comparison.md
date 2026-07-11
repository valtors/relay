# relay vs other MCP servers

## Quick comparison

| Feature | relay | Memory MCP | Fetch MCP | Screenshot MCP |
|---|---|---|---|---|
| Tools | 40 across 7 categories | 3-5 (memory only) | 1-2 (fetch only) | 1 (screenshot only) |
| Binary | Single Go binary | Node.js | Node.js | Node.js/Python |
| Runtime deps | None (static binary) | Node.js + npm deps | Node.js + npm deps | Node.js + browser |
| Agent memory | Yes (SQLite-backed) | Yes | No | No |
| Web fetch | Yes | No | Yes | No |
| Screenshot | Yes | No | No | Yes |
| Search | Yes | No | No | No |
| File ops | Yes | No | No | No |
| Multi-agent | Yes (coordination tools) | No | No | No |
| License | MIT | Varies | Varies | Varies |

## Why one binary instead of multiple servers

Running multiple MCP servers means:

- Multiple processes to manage and restart
- Multiple entries in your client config
- More memory usage (each Node.js process is 50-100 MB)
- Update each server separately
- Debug across multiple processes

Relay gives you all 40 tools in one process:

- One binary to install and update
- One entry in your client config
- ~20 MB memory usage total
- `relay doctor` to diagnose issues
- `relay upgrade` to update in place

## Config comparison

### Before: multiple MCP servers

```json
{
  "mcpServers": {
    "memory": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    },
    "fetch": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-fetch"]
    },
    "screenshot": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-screenshot"]
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"]
    }
  }
}
```

### After: just relay

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

That's it. Memory, fetch, screenshot, file ops, search, and 35 more tools, all from one config entry.

## Tool categories

| Category | Tools | What it does |
|---|---|---|
| Memory | 6 | Store, search, and manage agent memory with SQLite |
| Web | 8 | Fetch URLs, take screenshots, search the web |
| Files | 7 | Read, write, and manage files |
| System | 5 | Run commands, manage processes |
| Search | 4 | Full-text search across memory and files |
| Agents | 6 | Coordinate multiple agents |
| Config | 4 | Manage relay configuration |

## When to use relay

- You want one MCP server instead of many
- You need agent memory that persists across sessions
- You want web fetch, screenshots, and search in one place
- You are building multi-agent systems and need coordination
- You care about resource usage (one binary vs many Node processes)

## When to use something else

- You only need a single specific tool (a dedicated server may be simpler)
- You need a tool that relay does not provide yet
- You prefer Node.js and want to modify the server code

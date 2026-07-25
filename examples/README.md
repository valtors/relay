# relay examples

## quick start

```bash
# zero config - opens setup wizard
npx userelay
```

## with claude desktop

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

## file tools

```bash
# relay provides these MCP tools to your agent:
# - read_file: read any file
# - write_file: write files
# - list_files: list directory contents
# - search_files: search file contents
# - move_file: move or rename files
```

## web tools

relay gives your agent web access without leaving the terminal:

```
agent: "what's on hacker news?"
-> relay fetches https://news.ycombinator.com
-> returns clean markdown to the agent
-> agent summarizes
```

## pdf tools

```
agent: "read the pdf at /tmp/paper.pdf and summarize"
-> relay extracts text from the pdf
-> returns text to the agent
-> agent summarizes
```

## image tools

```
agent: "analyze the screenshot at /tmp/screen.png"
-> relay reads the image
-> returns metadata (dimensions, format, size)
```

## configuration

relay reads its config from `~/.relay/config.json`:

```json
{
  "allowedPaths": ["/home/user", "/tmp"],
  "webFetch": true,
  "maxFileSize": "10MB"
}
```

## multiple agents

relay works with any MCP-compatible agent:
- claude desktop
- cline
- goose
- codex
- any client that speaks MCP

just point the client at relay. it gets all tools in one server.

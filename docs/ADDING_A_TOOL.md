# Adding a Tool to Relay

This is the easiest first PR in Relay.

A new tool usually means:

- 1 new file in `tools/`
- 1 registration in `main.go`
- 1 or 2 focused tests
- 1 small update to the tool list test in `main_test.go`

If you can write one Go function, you can ship a tool here.

## Before you start

Run the repo once and make sure tests pass:

```bash
git clone https://github.com/valtors/relay
cd relay
go test ./...
```

If you already know what you want to build, open the `I want to add a tool` issue template first. That gives maintainers a chance to say "yes, good fit" before you spend time on it.

## Pick a tool category

Use a category to keep the scope tight.

- text: generate, rewrite, summarize, clean up text
- file: read, write, merge, or inspect local files
- data: parse, search, extract, or transform structured data
- workflow: orchestrate multiple steps or stitch together Relay outputs

This is just a planning label. Every tool still lives in `tools/`.

## Keep your first tool small

Good first tools:

- take 1 or 2 arguments
- return one clear result
- touch one part of the codebase
- do not need a brand new framework

Less good for a first PR:

- large multi-step orchestration
- new persistence models
- lots of config
- broad refactors

## Step 1 - copy the starter shape

Create a new file in `tools/`.

Example: `tools/echo_text.go`

```go
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func EchoText(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := strings.TrimSpace(req.GetString("input", ""))
	if input == "" {
		return mcp.NewToolResultError("input is required"), nil
	}

	out := fmt.Sprintf("echo: %s", input)
	return mcp.NewToolResultText(out), nil
}
```

Why this shape matters:

- the function lives in `tools`
- it takes `context.Context` and `mcp.CallToolRequest`
- validation errors come back as `mcp.NewToolResultError(...), nil`
- success comes back as `mcp.NewToolResultText(...)`

If your tool writes files into `./output/`, use helpers from `internal/state` like the existing tools do.

## Step 2 - register the tool in `main.go`

Find `buildServer()` in `main.go` and add your tool.

```go
s.AddTool(mcp.NewTool("echo_text",
	mcp.WithDescription("Echo text back to the caller. Good starter example."),
	mcp.WithString("input",
		mcp.Required(),
		mcp.Description("Text to echo"),
	),
), tools.EchoText)
```

A few naming tips:

- MCP tool name: lowercase snake case, like `echo_text`
- Go function name: exported CamelCase, like `EchoText`
- file name: match the tool name, like `echo_text.go`

## Step 3 - add focused tests

Create `tools/echo_text_test.go`.

```go
package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoText_InputRequired(t *testing.T) {
	res, err := EchoText(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError)
	assert.Contains(t, textOf(t, res), "input is required")
}

func TestEchoText_ReturnsEchoedText(t *testing.T) {
	res, err := EchoText(context.Background(), makeReq(map[string]any{
		"input": "hello",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.False(t, res.IsError)
	assert.Contains(t, textOf(t, res), "hello")
}
```

Useful test helpers already exist in `tools/request_approval_test.go`:

- `makeReq(...)` builds an MCP request
- `textOf(...)` reads text from a tool result
- `chdirTemp(t)` gives you a temp working directory when a tool reads or writes files

If your tool writes to `./output/`, call `chdirTemp(t)` first so your test stays isolated.

## Step 4 - update the server test

If you add a public tool, update `main_test.go` so the tool list expects it.

That test makes sure Relay actually registers every tool it ships.

## Step 5 - run checks

Before you open the PR:

```bash
gofmt -w .
go vet ./...
go test ./...
```

If you want a faster loop while building, it is fine to run a targeted test first and the full test suite before you push.

## Step 6 - try it with a real MCP client

You can test your new tool with a live client in two simple ways.

### Stdio

Point your MCP client at:

```json
{
  "command": "go",
  "args": ["run", "."]
}
```

### Streamable HTTP

Start Relay:

```bash
go run . -http -addr :8080
```

Then connect the client to `http://localhost:8080/mcp`.

If your tool makes real Anthropic calls, set `ANTHROPIC_API_KEY` in the shell or MCP client config first.

## Step 7 - open the PR

Keep the PR description short and useful:

- what the tool does
- why it belongs in Relay
- how you tested it
- sample input and output if that helps review

If this is your first PR, say so. Maintainers should give you a little extra context.

## Patterns that make review easy

- keep the tool small and single-purpose
- match existing file naming
- write clear descriptions for tool arguments
- prefer deterministic tests
- use `internal/state` helpers instead of rolling your own file handling
- keep code readable enough that it does not need comments

## Good files to copy from

Pick the closest shape and follow it:

- `tools/pm_plan.go` for a simple tool that reads input and writes one output file
- `tools/run_research.go` for an LLM-backed tool with upstream file checks
- `tools/request_approval.go` for validation-heavy request handling
- `tools/run_gtm.go` for a tool that coordinates more than one step

## Need help?

- Ask in [GitHub Discussions](https://github.com/valtors/relay/discussions)
- Open the `I want to add a tool` issue template
- Open a draft PR if you want feedback early

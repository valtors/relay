# Contributing to Relay

Thanks for wanting to help.

## Setup

```bash
git clone https://github.com/valtors/relay.git
cd relay
go test ./...
```

Requires Go 1.22+.

## Making Changes

1. Create a branch from `main`
2. Make your changes
3. Run `go test ./...` and `go vet ./...`
4. Open a PR

## PR Guidelines

- Keep PRs focused on a single change
- Write a clear title in imperative mood ("add X", "fix Y")
- Link related issues
- Include tests for new functionality

## What to Work On

Check [open issues](https://github.com/valtors/relay/issues) for tasks labeled `good first issue`.

Areas we need help with:
- Prompt improvements for agent stages
- Additional output formats
- New pipeline stages
- Documentation and examples
- Testing across MCP clients (Cursor, Windsurf, Continue)

## Code Style

- Run `gofmt` before committing
- No comments in code - use clear naming instead
- Follow Go conventions and idioms
- Keep functions short and focused

## Questions?

Open a [discussion](https://github.com/valtors/relay/discussions) or file an issue.

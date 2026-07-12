# Contributing to Relay

Thanks for your interest in contributing. Relay is a small, focused project. No contribution is too small.

## Quick Start

```bash
git clone https://github.com/valtors/relay.git
cd relay

# Build
go build -o relay .

# Run tests
go test -count=1 -short ./...

# Full pre-push check (run this before committing)
bash scripts/pre-push.sh
```

## Project Structure

```
main.go              # CLI entry point, command routing
cli_ui.go            # Terminal output helpers
doctor.go            # relay doctor diagnostic command
tools/
  registrations.go   # Tool registration (40 tools, 7 categories)
  file_tools.go      # File operations (read, write, hash, zip)
  image_tools.go     # Image manipulation (resize, crop, convert)
  web_tools.go       # Web fetch, screenshot, search
  text_tools.go      # Text extraction, templates, diff
  data_tools.go      # JSON, CSV, YAML, TOML parsing
  pdf_tools.go       # PDF extract, merge
  workflow_tools.go  # Workflow orchestration
internal/
  registry/          # Tool registry and lookup
  search/            # Fuzzy search
  ctxguard/           # Context cancellation guard
npm/                 # npm package (userelay)
scripts/
  pre-push.sh        # Pre-push verification script
```

## Adding a New Tool

1. Pick the right category file in `tools/` (or create a new one)
2. Write the tool function following the existing pattern
3. Register it in `tools/registrations.go`
4. Run `bash scripts/pre-push.sh` to verify
5. Open a PR

## Code Style

- Go stdlib first. Avoid external deps unless necessary.
- No comments in code unless the logic is non-obvious.
- Keep functions short and focused.
- Run `gofmt -w *.go` before committing.

## Scripts

| Command | Description |
|---|---|
| `bash scripts/pre-push.sh` | Full verification: gofmt, vet, test, build, unused imports |
| `go test -count=1 -short ./...` | Run tests (short mode) |
| `go vet ./...` | Static analysis |
| `gofmt -w *.go` | Format code |

## Good First Issues

Look for issues labeled `good first issue`. These are scoped for new contributors. If you get stuck, comment on the issue and we'll help.

## License

MIT. By contributing, you agree your contributions are licensed under the same terms.

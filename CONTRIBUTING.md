# Contributing to Relay

Thanks for checking out Relay.

If this is your first open source PR, this repo should feel easy to get into. If it feels confusing, that is on us, not you.

## Start here

- Want the easiest first PR? Read [docs/ADDING_A_TOOL.md](docs/ADDING_A_TOOL.md).
- Want a small task? Check issues labeled `good first issue` or `help wanted`.
- Want to ask before you code? Start a [GitHub Discussion](https://github.com/valtors/relay/discussions).

## The contributor experience we want

Relay should stay simple to try locally.

```bash
git clone https://github.com/valtors/relay
cd relay
go run .
```

No Docker. No setup maze. No local config scavenger hunt.

For contribution work, these should also just work:

```bash
go test ./...
go vet ./...
```

If a change makes local setup harder, it needs a very good reason.

## README first impression spec

The first screen of the README should do four jobs fast:

1. Say what Relay is in one sentence.
2. Show the fastest local start path.
3. Point contributors at one easy next step.
4. Show that maintainers are around and will reply.

Keep this near the top of the README:

- a `contributors welcome` badge
- a `good first issue` badge
- a short `New here?` block that links to:
  - `docs/ADDING_A_TOOL.md`
  - open `good first issue`s
  - GitHub Discussions
- one line that says when contributors should expect a reply

## Local setup

### What you need

- Git
- Go 1.22 or newer

### Run Relay locally

```bash
git clone https://github.com/valtors/relay
cd relay
go run .
```

### Run tests

```bash
go test ./...
```

### Format and quick checks

```bash
gofmt -w .
go vet ./...
```

### Test with a real MCP client during dev

You have two easy options.

#### Option 1: stdio

Point your MCP client at:

```json
{
  "command": "go",
  "args": ["run", "."]
}
```

Use this when you want the client to launch Relay for you.

#### Option 2: streamable HTTP

Start Relay yourself:

```bash
go run . -http -addr :8080
```

Then connect your MCP client to `http://localhost:8080/mcp`.

If you want to make real Anthropic calls, set `ANTHROPIC_API_KEY` in your shell or MCP client config. You do not need it for most test-only contribution work.

## Good first contribution ideas

- add a small tool
- improve a tool test
- tighten an error message
- improve client setup docs
- add an example prompt or sample workflow

If you are brand new, start with a tool. It is the cleanest path through the codebase.

## Add a tool

Relay keeps each MCP tool in its own file under `tools/`.

The usual flow is:

1. pick a small tool idea
2. add one file in `tools/`
3. register it in `main.go`
4. add focused tests
5. open a PR

The full guide is here: [docs/ADDING_A_TOOL.md](docs/ADDING_A_TOOL.md)

## Pull requests

Please keep PRs small and focused.

- one clear change per PR
- include tests when behavior changes
- update docs if the contributor or user flow changes
- link the issue if there is one
- drafts are welcome

You do not need to make it perfect before opening a PR. A small, clear PR beats a giant polished mystery every time.

## Review expectations

This is the part most projects get wrong. Here is the standard we should hold ourselves to:

- reply to new issues within 3 business days
- reply to new PRs within 3 business days
- if a PR cannot be reviewed yet, leave a short status comment
- no PR should sit for 7 days with zero maintainer response
- if feedback is blocking merge, be specific about what needs to change

If your PR has been quiet for a week, leave a short ping. That is normal and welcome.

## Communication

Use the channel that matches the problem:

- Bug reports: GitHub Issues
- Feature ideas: GitHub Issues
- "Is this a good first PR?" questions: GitHub Discussions
- "I am blocked and need a sanity check" questions: GitHub Discussions

If a discussion turns into concrete work, open or link an issue so the thread does not disappear.

## Recognition

We want contributors to feel seen.

- thank every merged contributor in the merge comment
- mention first-time contributors in the next release notes
- keep using the GitHub contributors page as the source of truth for names
- add a small Hall of Fame section to the README once the project has a steady contributor base

We are not using `CONTRIBUTORS.md` right now because it tends to go stale fast.

## Before you open a PR

Run:

```bash
gofmt -w .
go vet ./...
go test ./...
```

Then make sure your PR says:

- what changed
- why it changed
- how you tested it
- anything you want review help with

## Maintainer checklist

If you maintain Relay, this part matters just as much as the code:

- keep `good first issue` issues truly small and self-contained
- thank first-time contributors early, even before full review
- close the loop after merge
- do not let contribution threads die in silence

Fast, clear feedback is part of the product.

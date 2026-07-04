# Relay Zero-Friction Install UX Spec

## Goal

Make the first successful Relay experience happen in **under 2 minutes** with:

1. one install command,
2. one setup command,
3. one obvious first prompt,
4. no manual JSON editing.

Relay should feel like: **install, detect, configure, verify, try one thing, done.**

---

## Product Principles

1. **Default to success**: Relay should detect OS, arch, install path, and likely MCP clients automatically.
2. **No docs required**: the CLI should teach the user in the terminal as it goes.
3. **Never make users hand-edit JSON unless they choose to.**
4. **Show one screen at a time**: short, friendly, high-signal output only.
5. **Always leave the machine in a recoverable state**: create backups before patching config.
6. **Verify immediately**: every setup flow ends with a self-check and a “what to do next”.

---

## Primary Happy Path

### macOS / Linux

```bash
curl -fsSL https://relay.dev/install.sh | sh
relay
```

### Windows

```powershell
irm https://relay.dev/install.ps1 | iex
relay
```

### Expected experience

1. installer downloads the correct binary,
2. installs it into a usable PATH location,
3. prints `Run: relay`,
4. first `relay` launch opens guided setup,
5. Relay detects Claude Desktop / Cursor / Windsurf,
6. Relay patches config automatically,
7. Relay prints one example prompt to try,
8. user opens client, asks prompt, sees Relay work.

---

## 1. Install Methods (lowest friction first)

| Rank | Method | Command | Target user | Why it exists |
|---|---|---|---|---|
| 1 | Mac/Linux one-line installer | `curl -fsSL https://relay.dev/install.sh | sh` | Most users | Fastest path; no Go required |
| 2 | Windows one-line installer | `irm https://relay.dev/install.ps1 \| iex` | Most Windows users | Equivalent zero-friction path |
| 3 | Homebrew | `brew install valtors/tap/relay` | macOS power users | Familiar, trusted package manager |
| 4 | winget / Scoop | `winget install Valtors.Relay` or `scoop install relay` | Windows power users | Native Windows package flow |
| 5 | Go install | `go install github.com/valtors/relay@latest` | Go developers | Uses existing toolchain |
| 6 | Docker | `docker run --rm -it ghcr.io/valtors/relay:latest` | Users avoiding local install | Useful for testing and CI |
| 7 | GitHub Releases | manual download | Everyone else | Last-resort/manual fallback |

### 1.1 Mac/Linux one-line installer

#### Command

```bash
curl -fsSL https://relay.dev/install.sh | sh
```

#### Installer behavior

1. Detect OS and CPU architecture.
2. Resolve latest stable release.
3. Download checksum + archive.
4. Verify checksum.
5. Install to first writable preferred path:
   - `~/.local/bin`
   - `/usr/local/bin`
   - `/opt/homebrew/bin` on Apple Silicon if writable
6. Add path hint if install dir is not on PATH.
7. Print exactly one next step: `relay`

#### Mock terminal output

```text
$ curl -fsSL https://relay.dev/install.sh | sh

Relay installer
---------------
✓ Detected platform: macOS arm64
✓ Found latest release: v0.9.0
✓ Downloaded relay
✓ Verified checksum
✓ Installed to /opt/homebrew/bin/relay

Next step:
  relay
```

#### Failure output if PATH is missing

```text
✓ Installed to /Users/sam/.local/bin/relay

One more step to make `relay` available everywhere:
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
  source ~/.zshrc

Then run:
  relay
```

### 1.2 Windows one-line installer

#### Command

```powershell
irm https://relay.dev/install.ps1 | iex
```

#### Installer behavior

1. Detect x64 vs arm64.
2. Download signed `.zip`.
3. Verify checksum/signature.
4. Install to `%USERPROFILE%\AppData\Local\Programs\Relay\`.
5. Add install directory to user PATH if needed.
6. Print `Run: relay`.

#### Mock terminal output

```text
PS> irm https://relay.dev/install.ps1 | iex

Relay installer
---------------
✓ Detected platform: Windows x64
✓ Found latest release: v0.9.0
✓ Downloaded relay.exe
✓ Verified checksum
✓ Installed to C:\Users\sam\AppData\Local\Programs\Relay\relay.exe
✓ Added Relay to your user PATH

Next step:
  relay
```

### 1.3 Homebrew

#### Command

```bash
brew install valtors/tap/relay
```

#### Post-install message

```text
==> Relay installed.
Run `relay` to set up Claude Desktop, Cursor, or Windsurf.
```

#### UX note

Homebrew should remain a pure package install. Setup still begins on first `relay` launch.

### 1.4 winget / Scoop

#### Commands

```powershell
winget install Valtors.Relay
```

```powershell
scoop install relay
```

#### Post-install message

```text
Relay installed.
Run `relay` to connect it to your MCP client.
```

### 1.5 Go install

#### Command

```bash
go install github.com/valtors/relay@latest
```

#### Required UX improvement

After `go install`, the README and CLI must assume the user may not have `$GOPATH/bin` on PATH.

#### If `relay` is not found

```text
$ relay
zsh: command not found: relay
```

Relay docs and `go install` output should immediately offer:

```text
Relay was installed by Go, but your Go bin directory may not be on PATH.

Try:
  export PATH="$(go env GOPATH)/bin:$PATH"

Then run:
  relay

No Go? Use the one-line installer instead:
  curl -fsSL https://relay.dev/install.sh | sh
```

### 1.6 Docker

#### Command

```bash
docker run --rm -it ghcr.io/valtors/relay:latest
```

#### Positioning

Docker is **not** the default recommendation for desktop MCP clients. It is a fallback for:

- CI,
- locked-down environments,
- users who want zero host binary install.

#### UX note

`relay setup` should still emit a valid client config using a Docker command wrapper when the user chooses Docker mode.

Example generated MCP command:

```json
{
  "mcpServers": {
    "relay": {
      "command": "docker",
      "args": ["run", "--rm", "-i", "ghcr.io/valtors/relay:latest"]
    }
  }
}
```

### 1.7 GitHub Releases

#### UX requirements

Release assets must be named predictably:

- `relay_darwin_arm64.tar.gz`
- `relay_darwin_amd64.tar.gz`
- `relay_linux_arm64.tar.gz`
- `relay_linux_amd64.tar.gz`
- `relay_windows_arm64.zip`
- `relay_windows_amd64.zip`

Release page must show a copyable block:

```text
Downloaded Relay? Next step:
1. Put the binary on your PATH
2. Run `relay`
```

---

## 2. First-Run Experience

## Decision

When the user types `relay` with no prior setup, Relay should launch a **guided onboarding flow**.

### 2.1 What happens on `relay`

#### If Relay is already configured

Show concise status, then run normally.

```text
$ relay
Relay ready.
Configured clients: Claude Desktop, Cursor
Run `relay doctor` for diagnostics or start your MCP client.
```

#### If Relay is not configured yet

Launch setup automatically.

```text
$ relay

Welcome to Relay
It looks like this is your first run.

I can connect Relay to your MCP client automatically.

Detected on this machine:
  ✓ Claude Desktop
  ✓ Cursor
  – Windsurf not found

What would you like to do?
  1. Configure all detected clients
  2. Choose clients
  3. Print JSON snippet only
  4. Exit

Select an option [1]:
```

### 2.2 Should Relay auto-detect clients?

**Yes.** Always.

Detection sources:

- known config file locations,
- known app install directories,
- known process names,
- existing MCP config references.

### 2.3 Should Relay show an interactive menu?

**Yes, but only when needed.**

Rules:

1. **0 clients detected** → no menu, show install guidance + generic JSON snippet.
2. **1 client detected** → show a single confirmation prompt, not a full menu.
3. **2+ clients detected** → show menu.
4. **Non-interactive shell** → never prompt; print instructions and snippet.

#### 1-client mock output

```text
$ relay

Welcome to Relay
Found Claude Desktop.

Configure it now? [Y/n]:
```

### 2.4 Should Relay auto-configure clients?

**Yes, with explicit user consent.**

Rules:

- default answer should be the recommended safe action,
- always create a timestamped backup before patching,
- never overwrite a non-Relay entry,
- if `relay` entry already exists, offer update/keep/rename.

### 2.5 What output should users see?

Keep it to:

1. what Relay found,
2. what it changed,
3. whether verification passed,
4. exactly what to do next.

Avoid:

- raw JSON dumps unless requested,
- stack traces unless `--verbose`,
- more than one screen of text at a time.

### 2.6 How do they know it worked?

Success requires all of the following:

1. Relay binary runs.
2. Config file exists and contains a valid `relay` server entry.
3. Relay self-test passes.
4. User sees a concrete next action.

#### Success output

```text
✓ Relay installed correctly
✓ Claude Desktop configured
✓ Cursor configured
✓ Relay self-test passed

Next:
  1. Restart Claude Desktop and Cursor
  2. Open a project or attach a file
  3. Ask: Use Relay to count the words in README.md

Need help later?
  relay doctor
```

---

## 3. Auto-Configuration

## Command

```bash
relay setup
```

## Required behavior

`relay setup` must:

1. detect installed clients,
2. ask which to configure,
3. patch config automatically,
4. verify Relay can start,
5. print success + next steps.

### 3.1 Config file locations

> Relay should probe all known locations and prefer the canonical one for the installed platform.

#### Claude Desktop

| OS | Path |
|---|---|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

#### Cursor

| Scope | macOS / Linux | Windows |
|---|---|---|
| Global | `~/.cursor/mcp.json` | `%USERPROFILE%\.cursor\mcp.json` |
| Project | `<project>/.cursor/mcp.json` | `<project>\.cursor\mcp.json` |

Default choice: **global config** for fastest first-run success.

#### Windsurf

| OS | Path |
|---|---|
| macOS / Linux | `~/.codeium/windsurf/mcp_config.json` |
| Windows | `%USERPROFILE%\.codeium\windsurf\mcp_config.json` |

#### Generic

If no supported client is found, Relay prints a ready-to-paste JSON snippet.

### 3.2 Config patch strategy

For every target file:

1. create parent directories if missing,
2. create file if missing,
3. parse existing JSON,
4. create `mcpServers` object if missing,
5. insert or update only `mcpServers.relay`,
6. write pretty-printed JSON,
7. create backup: `<filename>.bak.<timestamp>`.

### 3.3 Canonical generated entry

#### Native binary

```json
{
  "mcpServers": {
    "relay": {
      "command": "C:\\Users\\sam\\AppData\\Local\\Programs\\Relay\\relay.exe",
      "args": [],
      "env": {}
    }
  }
}
```

On macOS/Linux, `command` should be the absolute installed binary path.

### 3.4 If a Relay entry already exists

Relay should show:

```text
Relay is already configured in Cursor:
  C:\Users\sam\.cursor\mcp.json

Current command:
  C:\old\path\relay.exe

Recommended command:
  C:\Users\sam\AppData\Local\Programs\Relay\relay.exe

What would you like to do?
  1. Update existing Relay entry
  2. Keep existing entry
  3. Add a second entry as relay-dev

Select an option [1]:
```

### 3.5 Verification flow

Verification has two layers.

#### Layer 1: immediate self-test

Run an internal smoke test:

```bash
relay doctor --self-test
```

Checks:

- binary is executable,
- JSON-RPC stdio server boots,
- advertised tool list loads,
- local Relay state directory is writable,
- optional localhost service port is available or reassigned.

#### Layer 2: client-specific verification

If the client supports live reload or can be launched safely:

- offer to restart/open it,
- verify the config file timestamp changed,
- instruct the user where to look in the client UI.

Example:

```text
Claude Desktop configured.
Restart Claude Desktop to load new MCP servers.
After restart, Relay should appear in Claude's available tools.
```

### 3.6 Suggested `relay setup` flow

```text
$ relay setup

Looking for MCP clients...
  ✓ Claude Desktop
  ✓ Cursor
  – Windsurf not found

Configure which clients?
  1. All detected clients
  2. Claude Desktop only
  3. Cursor only
  4. Print JSON snippet only

Select an option [1]:

Updating config...
  ✓ Backed up Claude Desktop config
  ✓ Patched Claude Desktop config
  ✓ Backed up Cursor config
  ✓ Patched Cursor config

Running Relay self-test...
  ✓ Relay starts correctly
  ✓ Tools loaded
  ✓ Local state directory writable

Setup complete.

Next:
  1. Restart Claude Desktop and Cursor
  2. Ask: Use Relay to count the words in README.md
```

---

## 4. First Task ("Hello World")

## Design goal

The first task should prove three things instantly:

1. Relay is connected,
2. Relay can use tools on local files,
3. the user understands what kind of things Relay can do.

## Canonical hello world

If the current working directory contains `README.md`, Relay should recommend:

```text
Use Relay to count the words in README.md
```

Why this is the right first task:

- short,
- concrete,
- visibly tool-driven,
- works without extra accounts or remote services,
- easy for the user to verify.

## Fallback prompt if no README exists

```text
Use Relay to list the files in my current folder and summarize what this project is.
```

## Client-specific guidance

### Claude Desktop

```text
Open Claude Desktop and ask:
  Use Relay to count the words in README.md

If Claude Desktop is not opened on a project, attach a folder or file first.
```

### Cursor / Windsurf

```text
Open any project and ask:
  Use Relay to count the words in README.md
```

## End-of-setup message

```text
Try Relay now:
  Use Relay to count the words in README.md

That should take one step and return a clear number.
```

---

## 5. Error Handling

## Principle

Every error message must do three things:

1. say what happened,
2. say what Relay did about it,
3. say the next best action.

### 5.1 Go is not installed

This only matters for the Go install path.

#### Bad experience

```text
go: command not found
```

#### Required Relay guidance

```text
Go is not installed, so `go install` will not work here.

Fastest option:
  curl -fsSL https://relay.dev/install.sh | sh

Windows:
  irm https://relay.dev/install.ps1 | iex

Or download a binary from:
  https://github.com/valtors/relay/releases/latest
```

### 5.2 Port 3000 is taken

If Relay uses a local helper service or dashboard on port 3000:

1. do **not** fail setup,
2. find the next free port automatically,
3. persist it to Relay config,
4. only mention it if relevant.

#### Output

```text
Port 3000 is already in use.
✓ Relay automatically switched to port 3001
```

No confirmation prompt needed.

### 5.3 Claude Desktop is not installed

#### Output

```text
Claude Desktop was not found on this machine.

You can still use Relay with:
  ✓ Cursor
  ✓ Windsurf
  ✓ any MCP client via JSON snippet

Print generic config now? [Y/n]:
```

### 5.4 Config file does not exist

Relay should create it automatically.

#### Output

```text
Cursor config not found.
✓ Created C:\Users\sam\.cursor\mcp.json
✓ Added Relay
```

### 5.5 Config file is invalid JSON

#### Output

```text
Cursor config exists but contains invalid JSON.

To keep your file safe, Relay did not overwrite it.
Backup created:
  C:\Users\sam\.cursor\mcp.json.broken.2026-07-04T193700

Choose one:
  1. Replace with a valid config containing Relay
  2. Print a JSON snippet for manual merge
  3. Exit
```

### 5.6 No supported clients detected

#### Output

```text
No supported MCP clients were detected.

Relay can still work with any MCP client.
Here is a generic config snippet:

{
  "mcpServers": {
    "relay": {
      "command": "/usr/local/bin/relay"
    }
  }
}

Known clients:
  - Claude Desktop
  - Cursor
  - Windsurf

Run `relay setup` again after installing one of them.
```

### 5.7 Relay binary moved after setup

`relay doctor` should detect stale paths.

#### Output

```text
Relay is configured in Claude Desktop, but the configured binary no longer exists:
  /usr/local/bin/relay

Found Relay here instead:
  /Users/sam/.local/bin/relay

Fix configs now? [Y/n]:
```

### 5.8 Permission denied writing config

#### Output

```text
Relay could not write:
  C:\Users\sam\AppData\Roaming\Claude\claude_desktop_config.json

Reason:
  Access denied

Try one of these:
  1. Re-run this command with permission to edit that file
  2. Print the JSON snippet and paste it manually
```

### 5.9 Client is installed but currently running

If hot reload is unsupported, Relay should warn gently.

```text
Claude Desktop is currently running.
Your config was updated, but Claude Desktop must be restarted before Relay appears.
```

---

## Interactive Setup Decision Tree

```text
User runs `relay`
|
+-- Is Relay already configured?
|   |
|   +-- Yes --> show "Relay ready" status --> exit
|   |
|   +-- No
|       |
|       +-- Interactive terminal?
|           |
|           +-- No --> print generic setup instructions + JSON snippet --> exit
|           |
|           +-- Yes
|               |
|               +-- Detect supported clients
|                   |
|                   +-- 0 found --> offer generic JSON snippet + client install hints
|                   |
|                   +-- 1 found --> ask "Configure now? [Y/n]"
|                   |
|                   +-- 2+ found --> show "all / choose / snippet / exit" menu
|                       |
|                       +-- Patch config(s)
|                       +-- Create backup(s)
|                       +-- Run self-test
|                       +-- Print first prompt
|                       +-- Done
```

---

## Generic JSON Snippet

Relay should always be able to print a copy-paste fallback.

### macOS / Linux

```json
{
  "mcpServers": {
    "relay": {
      "command": "/usr/local/bin/relay",
      "args": []
    }
  }
}
```

### Windows

```json
{
  "mcpServers": {
    "relay": {
      "command": "C:\\Users\\sam\\AppData\\Local\\Programs\\Relay\\relay.exe",
      "args": []
    }
  }
}
```

---

## Recommended Supporting Commands

These commands are not optional polish; they are part of the zero-friction UX.

### `relay setup`

Guided onboarding and config patching.

### `relay doctor`

Diagnostics for:

- binary location,
- PATH issues,
- detected clients,
- config validity,
- stale binary paths,
- port conflicts,
- self-test.

### `relay setup --print`

Print generic JSON without making changes.

### `relay setup --client claude`

Direct setup for scripts and advanced users.

### `relay setup --yes`

Non-interactive mode using safe defaults.

### `relay uninstall`

Optional, but valuable:

- remove Relay entries from known client configs,
- optionally remove installed binary,
- leave backups intact.

---

## UX Copy Guidelines

All Relay install/setup copy should sound:

- calm,
- competent,
- brief,
- optimistic.

Preferred style:

```text
✓ Found Cursor
✓ Added Relay
Next: restart Cursor
```

Avoid:

```text
Attempting to enumerate environment-specific editor integration assets...
```

---

## Final Recommendation

The best install flow is:

1. **one-line installer** per platform,
2. **`relay` launches setup automatically** on first run,
3. **auto-detect + consent-based auto-config** for supported clients,
4. **self-test + clear success message**,
5. **one tiny first prompt** that proves value immediately.

If a user ever has to read docs before seeing Relay work, the UX has failed.

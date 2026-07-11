import { Box, Text, useState, useEffect, SelectInput, html } from "../h.js";
import { BrailleSpinner, AnimatedBanner, GradientRule, KeyHint } from "../components.js";
import { existsSync, writeFileSync, mkdirSync, readFileSync } from "fs";
import { join } from "path";
import { homedir } from "os";

const editors = [
  {
    id: "claude",
    name: "Claude Desktop",
    icon: "🧠",
    detect: () => {
      const paths = [
        join(homedir(), "Library", "Application Support", "Claude", "claude_desktop_config.json"),
        join(homedir(), ".config", "Claude", "claude_desktop_config.json"),
        process.env.APPDATA ? join(process.env.APPDATA, "Claude", "claude_desktop_config.json") : null,
      ].filter(Boolean);
      return paths.find(existsSync) || null;
    },
  },
  {
    id: "cursor",
    name: "Cursor",
    icon: "✨",
    detect: () => {
      const paths = [
        join(homedir(), ".cursor", "mcp.json"),
        process.env.USERPROFILE ? join(process.env.USERPROFILE, ".cursor", "mcp.json") : null,
      ].filter(Boolean);
      return paths.find(existsSync) || null;
    },
  },
  {
    id: "vscode",
    name: "VS Code",
    icon: "📝",
    detect: () => {
      const p = join(homedir(), ".vscode", "mcp.json");
      return existsSync(p) ? p : null;
    },
  },
  {
    id: "windsurf",
    name: "Windsurf",
    icon: "🏄",
    detect: () => {
      const p = join(homedir(), ".windsurf", "mcp.json");
      return existsSync(p) ? p : null;
    },
  },
];

function isConfigured(configPath, editorId) {
  if (!configPath || !existsSync(configPath)) return false;
  try {
    const cfg = JSON.parse(readFileSync(configPath, "utf8"));
    if (editorId === "vscode") return !!cfg.servers?.relay;
    return !!cfg.mcpServers?.relay;
  } catch {
    return false;
  }
}

function writeEditorConfig(editorId, configPath, command) {
  const dir = join(configPath, "..");
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true });
  let cfg = {};
  if (existsSync(configPath)) {
    try {
      cfg = JSON.parse(readFileSync(configPath, "utf8"));
    } catch {
      cfg = {};
    }
  }
  if (editorId === "vscode") {
    cfg.servers = cfg.servers || {};
    cfg.servers.relay = { type: "stdio", command };
  } else {
    cfg.mcpServers = cfg.mcpServers || {};
    cfg.mcpServers.relay = { command };
  }
  writeFileSync(configPath, JSON.stringify(cfg, null, 2));
}

export function SetupWizard({ version, toolCount, binaryPath, onDone }) {
  const [phase, setPhase] = useState("intro");
  const [checkedEditors, setCheckedEditors] = useState([]);
  const [currentEditorIndex, setCurrentEditorIndex] = useState(0);
  const [progress, setProgress] = useState(0);
  const [log, setLog] = useState([]);
  const [results, setResults] = useState([]);

  useEffect(() => {
    if (phase !== "intro") return;
    const t = setTimeout(() => setPhase("detect"), 1800);
    return () => clearTimeout(t);
  }, [phase]);

  useEffect(() => {
    if (phase !== "detect") return;
    const found = editors
      .map((e) => ({ ...e, configPath: e.detect() }))
      .filter((e) => e.configPath);
    setCheckedEditors(found);
    setTimeout(() => {
      if (found.length === 0) setPhase("no-editors");
      else setPhase("download");
    }, 1200);
  }, [phase]);

  useEffect(() => {
    if (phase !== "download") return;
    let mounted = true;
    const interval = setInterval(() => {
      setProgress((p) => {
        const next = p + Math.random() * 15;
        if (next >= 100) {
          clearInterval(interval);
          if (mounted) setTimeout(() => setPhase("confirm"), 300);
          return 100;
        }
        return next;
      });
    }, 250);
    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, [phase]);

  if (phase === "intro") {
    return html`
      <${Box} flexDirection="column" alignItems="center" justifyContent="center" paddingTop=${2}>
        <${AnimatedBanner} text="RELAY" onDone=${() => {}} />
        <${Text} color="gray">Setting up your local MCP server</\Text>
      <//>
    `;
  }

  if (phase === "detect") {
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${BrailleSpinner} label="Detecting editors..." />
      <//>
    `;
  }

  if (phase === "no-editors") {
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${Text} color="yellow">No editors found on this machine.<//>
        <${Text} color="gray" marginTop=${1}>Install Claude Desktop, Cursor, VS Code, or Windsurf first.<//>
        <${Box} marginTop=${2}>
          <${SelectInput}
            items=${[
              { label: "Continue to menu", value: "done" },
              { label: "Retry detection", value: "retry" },
            ]}
            onSelect=${(item) => (item.value === "retry" ? setPhase("detect") : onDone())}
          />
        <//>
      <//>
    `;
  }

  if (phase === "download") {
    const bar = "█".repeat(Math.floor(progress / 5)) + "░".repeat(20 - Math.floor(progress / 5));
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${Text} color="cyan" bold>Downloading Relay v${version}<//>
        <${Text} color="gray">Installing Go binary for ${toolCount} tools<//>
        <${Box} marginTop=${1}>
          <${Text} color="cyan">[${bar}] ${Math.round(progress)}%<//>
        <//>
        <${GradientRule} width=${40} />
        <${Text} color="gray" marginTop=${1}>${binaryPath}<//>
      <//>
    `;
  }

  if (phase === "confirm") {
    const current = checkedEditors[currentEditorIndex];
    const already = isConfigured(current.configPath, current.id);
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${Text} color="cyan" bold>Setup editor ${currentEditorIndex + 1} of ${checkedEditors.length}<//>
        <${Text} color="white" marginTop=${1}>
          ${current.icon} ${current.name}
        <//>
        <${Text} color="gray">${current.configPath}<//>
        ${already
          ? html`<${Text} color="yellow" marginTop=${1}>⚠ Relay already configured. Overwrite?<//>`
          : html`<${Text} color="gray" marginTop=${1}>Add Relay to this editor?<//>`}
        <${Box} marginTop=${2}>
          <${SelectInput}
            items=${[
              { label: "Yes", value: "yes" },
              { label: "No", value: "no" },
            ]}
            onSelect=${(item) => {
              const res = [...results];
              const command = binaryPath || "relay";
              if (item.value === "yes") {
                try {
                  writeEditorConfig(current.id, current.configPath, command);
                  res.push({ name: current.name, success: true });
                } catch (err) {
                  res.push({ name: current.name, success: false, error: err.message });
                }
              } else {
                res.push({ name: current.name, skipped: true });
              }
              setResults(res);
              if (currentEditorIndex + 1 < checkedEditors.length) {
                setCurrentEditorIndex(currentEditorIndex + 1);
              } else {
                setPhase("done");
              }
            }}
          />
        <//>
      <//>
    `;
  }

  if (phase === "done") {
    const successCount = results.filter((r) => r.success).length;
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${Text} color="cyan" bold>✓ Setup complete<//>
        <${Text} color="gray">Configured ${successCount} editor${successCount === 1 ? "" : "s"}<//>
        ${results.map(
          (r, i) => html`
            <${Box} key=${i} marginTop=${1}>
              ${r.success
                ? html`<${Text} color="green">✓ ${r.name}<//>`
                : r.skipped
                ? html`<${Text} color="gray">⊘ ${r.name} (skipped)<//>`
                : html`<${Text} color="red">✗ ${r.name}: ${r.error}<//>`}
            <//>
          `
        )}
        <${Text} color="gray" marginTop=${2}>Restart your editor to start using Relay.<//>
        <${Box} marginTop=${1}>
          <${SelectInput}
            items=${[
              { label: "Open Relay menu", value: "menu" },
              { label: "Exit", value: "exit" },
            ]}
            onSelect=${(item) => {
              if (item.value === "exit") process.exit(0);
              onDone();
            }}
          />
        <//>
      <//>
    `;
  }

  return null;
}

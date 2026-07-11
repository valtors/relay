import { Box, Text, useState, useEffect, html, SelectInput } from "../h.js";
import { BrailleSpinner, GradientRule, KeyHint } from "../components.js";
import { execSync } from "child_process";
import { existsSync, writeFileSync, mkdirSync, readFileSync } from "fs";
import { join } from "path";
import { homedir } from "os";

function detectEditors() {
  const home = homedir();
  const editors = [];
  const claudePaths = [
    join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
    join(home, ".config", "Claude", "claude_desktop_config.json"),
    process.env.APPDATA ? join(process.env.APPDATA, "Claude", "claude_desktop_config.json") : null,
  ].filter(Boolean);
  const claudePath = claudePaths.find(existsSync);
  if (claudePath) editors.push({ name: "Claude Desktop", configPath: claudePath });

  const cursorPaths = [
    join(home, ".cursor", "mcp.json"),
    process.env.USERPROFILE ? join(process.env.USERPROFILE, ".cursor", "mcp.json") : null,
  ].filter(Boolean);
  const cursorPath = cursorPaths.find(existsSync);
  if (cursorPath) editors.push({ name: "Cursor", configPath: cursorPath });

  const vscodePath = join(home, ".vscode", "mcp.json");
  if (existsSync(vscodePath)) editors.push({ name: "VS Code", configPath: vscodePath });

  return editors;
}

function isRelayConfigured(configPath, editorName) {
  if (!existsSync(configPath)) return false;
  try {
    const config = JSON.parse(readFileSync(configPath, "utf8"));
    return !!(config.mcpServers?.relay || config.servers?.relay);
  } catch {
    return false;
  }
}

function writeConfig(editor, binaryPath) {
  const configDir = join(editor.configPath, "..");
  if (!existsSync(configDir)) mkdirSync(configDir, { recursive: true });
  let config = {};
  if (existsSync(editor.configPath)) {
    try {
      config = JSON.parse(readFileSync(editor.configPath, "utf8"));
    } catch {
      config = {};
    }
  }
  if (editor.name === "VS Code") {
    if (!config.servers) config.servers = {};
    config.servers.relay = { command: binaryPath };
  } else {
    if (!config.mcpServers) config.mcpServers = {};
    config.mcpServers.relay = { command: binaryPath };
  }
  writeFileSync(editor.configPath, JSON.stringify(config, null, 2), "utf8");
  return editor.configPath;
}

function resolveBinaryPath() {
  try {
    return execSync(process.platform === "win32" ? "where relay" : "which relay", { encoding: "utf8" }).trim().split("\n")[0];
  } catch {
    return null;
  }
}

export function InitWizard({ onDone }) {
  const [phase, setPhase] = useState("scanning");
  const [editors, setEditors] = useState([]);
  const [selectedEditor, setSelectedEditor] = useState(null);
  const [results, setResults] = useState(null);
  const [binaryResolved, setBinaryResolved] = useState(null);

  useEffect(() => {
    if (phase !== "scanning") return;
    const detected = detectEditors();
    const binary = resolveBinaryPath();
    setBinaryResolved(binary);
    setEditors(detected);
    setPhase(detected.length > 0 ? "select" : "none");
  }, [phase]);

  if (phase === "scanning") {
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${BrailleSpinner} label="Scanning for MCP-compatible editors..." />
      <//>
    `;
  }

  if (phase === "none") {
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${Text} color="yellow">No MCP-compatible editors detected.<//>
        <${Text} color="gray">Make sure Claude Desktop, Cursor, or VS Code is installed.<//>
        ${!binaryResolved ? html`<${Text} color="yellow" marginTop=${1}>! Relay binary not on PATH. Install with: npm i -g userelay<//>` : null}
        <${Box} marginTop=${2}>
          <${Text} color="gray">Press any key to return<//>
        <//>
      <//>
    `;
  }

  if (phase === "select") {
    const items = [
      ...editors.map((e) => ({
        label: `${e.name}${isRelayConfigured(e.configPath, e.name) ? " (already configured)" : ""}`,
        value: e.name,
      })),
      { label: "Configure all detected editors", value: "all" },
      { label: "Back to menu", value: "back" },
    ];
    return html`
      <${Box} flexDirection="column" paddingTop=${1}>
        <${Text} color="cyan" bold>Detected editors:<//>
        ${!binaryResolved ? html`<${Text} color="yellow" marginTop=${1}>! Relay binary not on PATH — config will use "relay" as command<//>` : null}
        <${Box} marginTop=${1} marginBottom=${1}>
          <${KeyHint} hints=${["↑↓ navigate", "Enter select", "Ctrl+C quit"]} />
        <//>
        <${Box}>
          <${SelectInput} items=${items} onSelect=${(item) => {
            if (item.value === "back") return onDone();
            if (item.value === "all") {
              setSelectedEditor("all");
              setPhase("confirm");
            } else {
              const ed = editors.find((e) => e.name === item.value);
              setSelectedEditor(ed);
              setPhase("confirm");
            }
          }} />
        <//>
      <//>
    `;
  }

  if (phase === "confirm") {
    const targetLabel = selectedEditor === "all"
      ? editors.map((e) => e.name).join(", ")
      : selectedEditor.name;
    const targetPath = selectedEditor === "all"
      ? editors.map((e) => e.configPath).join("\n  ")
      : selectedEditor.configPath;
    const alreadyConfigured = selectedEditor === "all"
      ? editors.some((e) => isRelayConfigured(e.configPath, e.name))
      : isRelayConfigured(selectedEditor.configPath, selectedEditor.name);
    return html`
      <${Box} flexDirection="column" paddingTop=${1}>
        <${Text} color="cyan" bold>Confirm configuration<//>
        <${Box} marginTop=${1}>
          <${Text} color="white">Write Relay MCP config to: ${targetLabel}<//>
        <//>
        <${Box} marginTop=${1}>
          <${Text} color="gray">${targetPath}<//>
        <//>
        ${alreadyConfigured ? html`<${Text} color="yellow" marginTop=${1}>! Relay is already configured — this will overwrite the existing entry<//>` : null}
        <${GradientRule} width=${40} />
        <${Text} color="gray" marginTop=${1}>Add config?<//>
        <${Box}>
          <${SelectInput}
            items=${[
              { label: "Yes, write config", value: "yes" },
              { label: "Cancel", value: "cancel" },
            ]}
            onSelect=${(item) => {
              if (item.value === "cancel") {
                setPhase("select");
              } else {
                setPhase("writing");
              }
            }}
          />
        <//>
      <//>
    `;
  }

  if (phase === "writing") {
    const binaryPath = binaryResolved || "relay";
    if (selectedEditor === "all") {
      const writeResults = editors.map((e) => {
        try {
          const p = writeConfig(e, binaryPath);
          return { name: e.name, path: p, success: true };
        } catch (err) {
          return { name: e.name, error: err.message, success: false };
        }
      });
      setResults(writeResults);
      setPhase("done");
      return null;
    }
    try {
      const p = writeConfig(selectedEditor, binaryPath);
      setResults([{ name: selectedEditor.name, path: p, success: true }]);
      setPhase("done");
    } catch (err) {
      setResults([{ name: selectedEditor.name, error: err.message, success: false }]);
      setPhase("error");
    }
    return null;
  }

  if (phase === "done") {
    return html`
      <${Box} flexDirection="column" paddingTop=${2} alignItems="center">
        <${Text} color="cyan" bold>+ Done<//>
        ${results.map((r, i) => html`
          <${Box} key=${i}>
            ${r.success
              ? html`<${Text} color="green">+ ${r.name}: ${r.path}<//>`
              : html`<${Text} color="red">- ${r.name}: ${r.error}<//>`}
          <//>
        `)}
        <${Text} color="gray" marginTop=${1}>Restart your editor to pick up the new MCP server.<//>
        <${Box} marginTop=${1}>
          <${Text} color="gray">Press any key to return<//>
        <//>
      <//>
    `;
  }

  return html`
    <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
      <${Text} color="red">Failed to write config. Check file permissions.<//>
      <${Box} marginTop=${1}>
        <${Text} color="gray">Press any key to return<//>
      <//>
    <//>
  `;
}

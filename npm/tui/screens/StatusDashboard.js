import { Box, Text, useState, useEffect, html } from "../h.js";
import { GradientRule, Divider, KeyHint } from "../components.js";
import { SelectInput } from "../h.js";
import { existsSync, readFileSync } from "fs";
import { join } from "path";
import { homedir } from "os";
import { execSync } from "child_process";
function checkBinary() {
  try {
    const p = execSync("which relay || where relay", { encoding: "utf8" }).trim().split("\n")[0];
    return { found: true, path: p };
  } catch {
    return { found: false, path: "not on PATH" };
  }
}
function checkConfigs() {
  const home = homedir();
  const configs = [
    {
      name: "Claude Desktop",
      path: join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
    },
    {
      name: "Cursor",
      path: join(home, ".cursor", "mcp.json"),
    },
    {
      name: "VS Code",
      path: join(home, ".vscode", "mcp.json"),
    },
  ];
  return configs.map((c) => {
    if (!existsSync(c.path)) return { ...c, exists: false, hasRelay: false };
    try {
      const content = JSON.parse(readFileSync(c.path, "utf8"));
      const hasRelay = !!(content.mcpServers?.relay || content.servers?.relay);
      return { ...c, exists: true, hasRelay };
    } catch {
      return { ...c, exists: true, hasRelay: false };
    }
  });
}
export function StatusDashboard({ version, toolCount, binaryPath, onDone }) {
  const [binaryStatus, setBinaryStatus] = useState(null);
  const [configStatus, setConfigStatus] = useState([]);
  useEffect(() => {
    setBinaryStatus(checkBinary());
    setConfigStatus(checkConfigs());
  }, []);
  return html`
    <${Box} flexDirection="column" paddingTop=${2} paddingLeft=${2}>
      <${Text} color="cyan" bold>RELAY STATUS<//>
      <${GradientRule} width=${36} />
      <${Box} flexDirection="column" marginTop=${1}>
        <${Box}>
          <${Text} color="gray">Version      <//>
          <${Text} color="white">v${version}<//>
        <//>
        <${Box}>
          <${Text} color="gray">Tools        <//>
          <${Text} color="white">${toolCount} registered (7 categories)<//>
        <//>
        <${Box}>
          <${Text} color="gray">Binary       <//>
          ${binaryStatus
            ? (binaryStatus.found
              ? html`<${Text} color="cyan">${binaryStatus.path}<
              : html`<${Text} color="yellow">not on PATH<//>`)
            : html`<${Text} color="gray">checking...<//>`}
        <
      <
      <${Divider} width=${36} />
      <${Text} color="cyan" bold>Editor configs<
      <${Box} flexDirection="column" marginTop=${1}>
        ${configStatus.map((c, i) => html`
          <${Box} key=${i}>
            <${Text} color="gray">${c.name.padEnd(14)}<//>
            ${!c.exists
              ? html`<${Text} color="gray">not found<
              : c.hasRelay
              ? html`<${Text} color="cyan">configured ✓<//>`
              : html`<${Text} color="yellow">exists, no relay<//>`}
          <
        `)}
      <//>
      <${Box} marginTop=${2}>
        <${SelectInput}
          items=${[{ label: "Back to menu", value: "back" }]}
          onSelect=${() => onDone()}
        />
      <//>
      <${KeyHint} hints=${["Enter select", "Ctrl+C quit"]} />
    <//>
  `;
}

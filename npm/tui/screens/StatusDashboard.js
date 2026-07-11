import { Box, Text, useState, useEffect, useInput, html } from "../h.js";
import { BrailleSpinner, Divider } from "../components.js";
import { spawnSync } from "child_process";
import { existsSync } from "fs";
import { homedir } from "os";
import { join } from "path";

const CHECK_INTERVAL = 2000;

export function StatusDashboard({ version, toolCount, binaryPath, onDone }) {
  const [binaryStatus, setBinaryStatus] = useState(null);
  const [configStatus, setConfigStatus] = useState([]);
  const [now, setNow] = useState(new Date());

  const checkBinary = () => {
    if (existsSync(binaryPath)) {
      return { found: true, path: binaryPath };
    }
    try {
      const result = spawnSync(binaryPath, ["version"], { encoding: "utf8", timeout: 3000 });
      if (result.status === 0) {
        return { found: true, path: binaryPath };
      }
    } catch {}
    return { found: false };
  };

  const checkEditors = () => {
    const editors = [
      { name: "Claude", paths: [
        join(homedir(), "Library/Application Support/Claude/claude_desktop_config.json"),
        join(homedir(), ".config/Claude/claude_desktop_config.json"),
      ]},
      { name: "Cursor", paths: [join(homedir(), ".cursor/mcp.json")] },
      { name: "VS Code", paths: [join(homedir(), ".vscode/mcp.json")] },
    ];
    return editors.map((e) => {
      const found = e.paths.filter(existsSync);
      return { ...e, found };
    });
  };

  const refresh = () => {
    setBinaryStatus(checkBinary());
    setConfigStatus(checkEditors());
    setNow(new Date());
  };

  useInput(() => onDone());

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, CHECK_INTERVAL);
    return () => clearInterval(interval);
  }, []);

  if (!binaryStatus) {
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${BrailleSpinner} label="Checking Relay status..." />
      <//>
    `;
  }

  return html`
    <${Box} flexDirection="column" paddingTop=${1}>
      <${Text} color="cyan" bold>Relay status<//>
      <${Box} flexDirection="column" marginTop=${1}>
        <${Box}>
          <${Text} color="gray">Version      <//>
          <${Text} color="white">${version}<//>
        <//>
        <${Box}>
          <${Text} color="gray">Tools        <//>
          <${Text} color="white">${toolCount} registered (7 categories)<//>
        <//>
        <${Box}>
          <${Text} color="gray">Binary       <//>
          ${binaryStatus.found
            ? html`<${Text} color="cyan">${binaryStatus.path}<//>`
            : html`<${Text} color="yellow">not on PATH<//>`}
        <//>
      <//>
      <${Divider} width=${36} />
      <${Text} color="cyan" bold>Editor configs<//>
      <${Box} flexDirection="column" marginTop=${1}>
        ${configStatus.map((c, i) => html`
          <${Box} key=${i}>
            <${Text} color="gray">${c.name.padEnd(14)}<//>
            ${c.found.length > 0
              ? html`<${Text} color="cyan">found<//>`
              : html`<${Text} color="gray">not found<//>`}
          <//>
        `)}
      <//>
      <${Divider} width=${36} />
      <${Text} color="gray" marginTop=${1}>
        Last updated: ${now.toLocaleTimeString()}
      <//>
      <${Text} color="gray">Press any key to return...<//>
    <//>
  `;
}

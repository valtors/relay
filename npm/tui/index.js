import { render, Box, Text, useApp, useInput, useState, useEffect, html } from "./h.js";
import { LaunchScreen } from "./screens/LaunchScreen.js";
import { SetupWizard } from "./screens/SetupWizard.js";
import { InitWizard } from "./screens/InitWizard.js";
import { ToolsBrowser } from "./screens/ToolsBrowser.js";
import { StatusDashboard } from "./screens/StatusDashboard.js";
const CATEGORY_COUNT = 7;
function App({ version, toolCount, binaryPath, onExit, onStartServer, startScreen = "launch" }) {
  const [screen, setScreen] = useState(startScreen); 
  if (screen === "start" || screen === "start-http") {
    const mode = screen === "start-http" ? "http" : "stdio";
    if (onStartServer) {
      onStartServer(mode);
    }
    return html`
      <${Box} flexDirection="column" alignItems="center" paddingTop=${2}>
        <${Text} color="cyan">
          Starting Relay MCP server (${mode === "http" ? "HTTP" : "stdio"})...
        <//>
        <${Text} color="gray" marginTop=${1}>
          Handing off to the Go binary.
        <//>
      <//>
    `;
  }
  if (screen === "launch") {
    return html`<${LaunchScreen}
      version=${version}
      toolCount=${toolCount}
      categoryCount=${CATEGORY_COUNT}
      onSelect=${(value) => {
        if (value === "quit") {
          onExit();
        } else {
          setScreen(value);
        }
      }}
    />`;
  }
  if (screen === "setup") {
    return html`<${SetupWizard}
      version=${version}
      toolCount=${toolCount}
      binaryPath=${binaryPath}
      onDone=${() => setScreen("launch")}
    />`;
  }
  if (screen === "init") {
    return html`<${InitWizard} onDone=${() => setScreen("launch")} />`;
  }
  if (screen === "tools") {
    return html`<${ToolsBrowser} onDone=${() => setScreen("launch")} />`;
  }
  if (screen === "status") {
    return html`<${StatusDashboard}
      version=${version}
      toolCount=${toolCount}
      binaryPath=${binaryPath}
      onDone=${() => setScreen("launch")}
    />`;
  }
  return html`
    <${Box}>
      <${Text} color="red">Unknown screen: ${screen}<//>
    <//>
  `;
}
export function startTUI(opts = {}) {
  const {
    version = "dev",
    toolCount = 40,
    binaryPath = "relay",
    startScreen = "launch",
    onStartServer,
  } = opts;
  const { waitUntilExit, unmount } = render(
    html`<${App}
      version=${version}
      toolCount=${toolCount}
      binaryPath=${binaryPath}
      startScreen=${startScreen}
      onExit=${() => {
        unmount();
        process.exit(0);
      }}
      onStartServer=${onStartServer}
    />`,
    {
      exitOnCtrlC: true,
    }
  );
  return { waitUntilExit };
}

import { Box, Text, useEffect, useState, SelectInput, useInput, html } from "../h.js";
import { AnimatedBanner, GradientRule, KeyHint } from "../components.js";
export function LaunchScreen({ version, toolCount, categoryCount, onSelect }) {
  const [showBanner, setShowBanner] = useState(true);
  if (showBanner) {
    return html`
      <${Box} flexDirection="column" alignItems="center" justifyContent="center" paddingTop=${2}>
        <${AnimatedBanner}
          text="RELAY"
          onDone=${() => setShowBanner(false)}
        />
      <//>
    `;
  }
  const items = [
    { label: "Start MCP server (stdio)", value: "start" },
    { label: "Start MCP server (HTTP)", value: "start-http" },
    { label: "Initialize — detect & configure editors", value: "init" },
    { label: "Browse tools", value: "tools" },
    { label: "Status", value: "status" },
    { label: "Quit", value: "quit" },
  ];
  const infoLine = `relay v${version}  ·  ${toolCount} tools  ·  ${categoryCount} categories`;
  return html`
    <${Box} flexDirection="column" alignItems="center" paddingTop=${1}>
      <${Text} color="cyan" bold>${infoLine}<//>
      <${GradientRule} width=${Math.max(infoLine.length, 36)} />
      <${Box} marginTop=${1}>
        <${SelectInput} items=${items} onSelect=${(item) => onSelect(item.value)} />
      <//>
      <${KeyHint} hints=${["↑↓ navigate", "Enter select", "Ctrl+C quit"]} />
    <//>
  `;
}

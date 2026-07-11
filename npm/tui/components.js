import { Box, Text, useState, useEffect, html } from "./h.js";
import gradient from "gradient-string";
import { spinners as unicodeSpinners } from "unicode-animations";
import { animateShimmer } from "./shimmer.js";

const SPINNER_FRAMES = unicodeSpinners.braille.frames;
const SPINNER_INTERVAL = 80; 

export function BrailleSpinner({ label = "Loading" }) {
  const [frame, setFrame] = useState(0);
  useEffect(() => {
    const timer = setInterval(() => {
      setFrame((f) => (f + 1) % SPINNER_FRAMES.length);
    }, SPINNER_INTERVAL);
    return () => clearInterval(timer);
  }, []);
  return html`
    <${Box}>
      <${Text} color="cyan">${SPINNER_FRAMES[frame]}<//>
      <${Text}> ${label}<//>
    <//>
  `;
}

export function GradientRule({ width = 40, char = "─" }) {
  const line = char.repeat(width);
  const gradientLine = gradient(["#00d4ff", "#1e90ff"])(line);
  return html`
    <${Box}>
      <${Text}>${gradientLine}<//>
    <//>
  `;
}

export function Banner({ version = "dev", toolCount = 40 }) {
  const title = gradient(["#00d4ff", "#9b5de5"])("RELAY");
  const subtitle = `v${version}  ·  ${toolCount} tools  ·  MCP server`;
  return html`
    <${Box} flexDirection="column" alignItems="center">
      <${Text}>${title}<//>
      <${Text} color="gray">${subtitle}<//>
    <//>
  `;
}

export function AnimatedBanner({ text = "RELAY", onDone }) {
  const [frame, setFrame] = useState("");
  useEffect(() => {
    const cleanup = animateShimmer(text, (rendered) => {
      setFrame(rendered);
    }, { interval: 60 });
    const doneTimer = setTimeout(() => {
      cleanup();
      if (onDone) onDone();
    }, 1500);
    return () => {
      cleanup();
      clearTimeout(doneTimer);
    };
  }, []);
  if (!frame) return null;
  return html`
    <${Box} flexDirection="column" alignItems="center" marginBottom=${1}>
      <${Text}>${frame}<//>
    <//>
  `;
}

export function ToolList({ tools }) {
  const categories = Object.keys(tools);
  return html`
    <${Box} flexDirection="column">
      ${categories.map((cat) => html`
        <${Box} flexDirection="column" key=${cat} marginBottom=${1}>
          <${Text} color="cyan" bold>${cat.toUpperCase()}<//>
          ${tools[cat].map((tool) => html`
            <${Box} key=${tool.name}>
              <${Text} color="gray">  └ <//>
              <${Text} color="white" bold>${tool.name.padEnd(24)}<//>
              <${Text} color="gray">${tool.description}<//>
            <//>
          `)}
        <//>
      `)}
    <//>
  `;
}

export function StatusBadge({ status = "ready" }) {
  const colors = {
    ready: "cyan",
    running: "cyan",
    error: "red",
    waiting: "yellow",
  };
  const color = colors[status] || "gray";
  return html`
    <${Box}>
      <${Text} color=${color} bold>[${status.toUpperCase()}]<//>
    <//>
  `;
}

export function KeyHint({ hints = [] }) {
  return html`
    <${Box} marginTop=${1}>
      ${hints.map((h, i) => html`
        <${Text} color="gray" key=${i}>
          ${i > 0 ? "  " : ""}${h}
        <//>
      `)}
    <//>
  `;
}

export function Divider({ width = 40 }) {
  return html`
    <${Box}>
      <${Text} color="gray">${"─".repeat(width)}<//>
    <//>
  `;
}

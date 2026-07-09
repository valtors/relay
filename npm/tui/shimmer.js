import cfonts from "cfonts";
const BASE_COLOR = { r: 30, g: 144, b: 255 };      
const HIGHLIGHT_COLOR = { r: 220, g: 240, b: 255 }; 
const SHIMMER_WIDTH = 8; 
export function stripAnsi(str) {
  return str.replace(/\x1b\[[0-9;]*m/g, "");
}
export function renderWordmark(text, opts = {}) {
  const output = cfonts.render(text, {
    font: opts.font || "block",
    align: opts.align || "left",
    colors: ["#ffffff"],
    background: "transparent",
    lineHeight: opts.lineHeight || 1,
    space: opts.space ?? true,
    maxLength: opts.maxLength || 0,
    gradient: false,
    independentGradient: false,
    transitionGradient: false,
    rawMode: true,
    env: "node",
  });
  if (!output || !output.string) return [];
  return output.string
    .split("\n")
    .map((line) => stripAnsi(line.replace(/\r/g, "")))
    .filter((line) => line.length > 0);
}
export function buildGrid(lines) {
  return lines.map((line, row) =>
    Array.from(line).map((char, col) => ({ char, row, col }))
  );
}
export function lerpColor(a, b, t) {
  return {
    r: Math.round(a.r + (b.r - a.r) * t),
    g: Math.round(a.g + (b.g - a.g) * t),
    b: Math.round(a.b + (b.b - a.b) * t),
  };
}
export function rgbToAnsi(color) {
  return `\x1b[38;2;${color.r};${color.g};${color.b}m`;
}
export function renderFrame(grid, offset, width = SHIMMER_WIDTH) {
  const lines = grid.map((row) => {
    const chars = row.map((cell) => {
      if (cell.char === " ") return " ";
      const pos = cell.col + cell.row;
      const distance = pos - offset;
      if (distance >= 0 && distance < width) {
        const t = 1 - distance / width;
        const color = lerpColor(BASE_COLOR, HIGHLIGHT_COLOR, t);
        return `${rgbToAnsi(color)}${cell.char}\x1b[0m`;
      }
      return `${rgbToAnsi(BASE_COLOR)}${cell.char}\x1b[0m`;
    });
    return chars.join("");
  });
  return lines.join("\n");
}
export function getSweepRange(grid) {
  if (!grid.length) return 0;
  const maxCol = Math.max(...grid.map((row) => row.length));
  return maxCol + grid.length + SHIMMER_WIDTH;
}
export function animateShimmer(text, onFrame, opts = {}) {
  const interval = opts.interval || 60;
  const rawLines = renderWordmark(text, opts);
  const grid = buildGrid(rawLines);
  const sweepRange = getSweepRange(grid);
  let offset = -SHIMMER_WIDTH;
  onFrame(renderFrame(grid, offset));
  const timer = setInterval(() => {
    offset += 1;
    if (offset > sweepRange) offset = -SHIMMER_WIDTH;
    onFrame(renderFrame(grid, offset));
  }, interval);
  return () => clearInterval(timer);
}

import { test, describe } from "node:test";
import assert from "node:assert";
describe("shimmer", () => {
  test("renderWordmark returns plain text lines", async () => {
    const { renderWordmark, stripAnsi } = await import("../shimmer.js");
    const lines = renderWordmark("RELAY");
    assert.ok(lines.length > 0, "should return lines");
    assert.ok(lines.length >= 3, "block font should have multiple lines");
    lines.forEach((line) => {
      assert.ok(!line.includes("\x1b["), "lines should not have ANSI codes");
    });
  });
  test("stripAnsi removes escape codes", async () => {
    const { stripAnsi } = await import("../shimmer.js");
    const input = "\x1b[38;2;30;144;255mhello\x1b[0m world";
    assert.strictEqual(stripAnsi(input), "hello world");
  });
  test("buildGrid creates 2D array of cells", async () => {
    const { buildGrid } = await import("../shimmer.js");
    const lines = ["AB", "CD"];
    const grid = buildGrid(lines);
    assert.strictEqual(grid.length, 2);
    assert.strictEqual(grid[0].length, 2);
    assert.strictEqual(grid[0][0].char, "A");
    assert.strictEqual(grid[1][1].char, "D");
    assert.strictEqual(grid[0][0].row, 0);
    assert.strictEqual(grid[0][0].col, 0);
  });
  test("lerpColor interpolates between colors", async () => {
    const { lerpColor } = await import("../shimmer.js");
    const a = { r: 0, g: 0, b: 0 };
    const b = { r: 100, g: 100, b: 100 };
    const mid = lerpColor(a, b, 0.5);
    assert.strictEqual(mid.r, 50);
    assert.strictEqual(mid.g, 50);
    assert.strictEqual(mid.b, 50);
  });
  test("rgbToAnsi produces truecolor escape", async () => {
    const { rgbToAnsi } = await import("../shimmer.js");
    const ansi = rgbToAnsi({ r: 30, g: 144, b: 255 });
    assert.strictEqual(ansi, "\x1b[38;2;30;144;255m");
  });
  test("renderFrame produces ANSI-colored output", async () => {
    const { renderWordmark, buildGrid, renderFrame } = await import("../shimmer.js");
    const lines = renderWordmark("HI");
    const grid = buildGrid(lines);
    const frame = renderFrame(grid, 0);
    assert.ok(frame.includes("\x1b["), "frame should have ANSI codes");
    assert.ok(frame.includes("\n"), "frame should have line breaks");
  });
  test("getSweepRange returns positive number", async () => {
    const { renderWordmark, buildGrid, getSweepRange } = await import("../shimmer.js");
    const lines = renderWordmark("RELAY");
    const grid = buildGrid(lines);
    const range = getSweepRange(grid);
    assert.ok(range > 0, "sweep range should be positive");
  });
  test("animateShimmer calls onFrame and cleanup stops it", async () => {
    const { animateShimmer } = await import("../shimmer.js");
    let frameCount = 0;
    const cleanup = animateShimmer("RELAY", () => {
      frameCount++;
    }, { interval: 20 });
    await new Promise((resolve) => setTimeout(resolve, 100));
    cleanup();
    const countAtStop = frameCount;
    await new Promise((resolve) => setTimeout(resolve, 100));
    assert.strictEqual(frameCount, countAtStop, "animation should stop after cleanup");
    assert.ok(frameCount >= 2, "should have at least 2 frames");
  });
});
describe("h.js", () => {
  test("html template produces React elements", async () => {
    const { html, Box, Text } = await import("../h.js");
    const el = html`<${Box} flexDirection="column">
      <${Text} color="cyan">Test<//>
    <//>`;
    assert.ok(el.$$typeof, "should be a React element");
    assert.strictEqual(el.type, Box);
  });
  test("numeric props are passed as numbers", async () => {
    const { html, Box, Text } = await import("../h.js");
    const el = html`<${Box} marginTop=${1}><${Text}>test<//><//>`;
    assert.strictEqual(el.props.marginTop, 1);
    assert.strictEqual(typeof el.props.marginTop, "number");
  });
});
describe("components", () => {
  test("BrailleSpinner produces React element", async () => {
    const { html } = await import("../h.js");
    const { BrailleSpinner } = await import("../components.js");
    const el = html`<${BrailleSpinner} label="Loading" />`;
    assert.ok(el.$$typeof, "should be a React element");
  });
  test("Banner produces React element with version", async () => {
    const { html } = await import("../h.js");
    const { Banner } = await import("../components.js");
    const el = html`<${Banner} version="0.3.0" toolCount=${40} />`;
    assert.ok(el.$$typeof, "should be a React element");
  });
  test("ToolList produces React element", async () => {
    const { html } = await import("../h.js");
    const { ToolList } = await import("../components.js");
    const tools = {
      "File (7)": [
        { name: "file_read", description: "Read file" },
      ],
    };
    const el = html`<${ToolList} tools=${tools} />`;
    assert.ok(el.$$typeof, "should be a React element");
  });
  test("StatusBadge produces React element", async () => {
    const { html } = await import("../h.js");
    const { StatusBadge } = await import("../components.js");
    const el = html`<${StatusBadge} status="ready" />`;
    assert.ok(el.$$typeof, "should be a React element");
  });
});

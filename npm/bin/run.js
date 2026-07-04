#!/usr/bin/env node

// Thin wrapper that runs the relay binary
// Usage: npx relay-mcp [args...]

const { execFileSync } = require("child_process");
const path = require("path");

const BIN_NAME = process.platform === "win32" ? "relay.exe" : "relay";
const BIN_PATH = path.join(__dirname, BIN_NAME);

try {
  execFileSync(BIN_PATH, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  if (e.status !== null) {
    process.exit(e.status);
  }
  console.error("could not run relay. try: npx relay-mcp install");
  process.exit(1);
}

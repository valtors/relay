#!/usr/bin/env node
const { spawnSync } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");
const {
  PACKAGE_VERSION,
  fileExists,
  getCacheBinaryPath,
  getFallbackBinaryPath,
  isCachedVersionInstalled,
} = require("./lib");
const { install } = require("./install");

function resolveBinaryPath() {
  if (isCachedVersionInstalled(PACKAGE_VERSION)) {
    return getCacheBinaryPath();
  }
  const fallbackPath = getFallbackBinaryPath();
  return fileExists(fallbackPath) ? fallbackPath : "";
}

function shouldLaunchTUI(args) {
  if (args.length === 0) {
    return true;
  }
  const interactive = process.stdin.isTTY || process.stdout.isTTY || process.stderr.isTTY;
  if (args[0] === "tui" && interactive) {
    return true;
  }
  return false;
}

const setupFlagPath = path.join(os.homedir(), ".config", "relay", "setup-complete");

function isFirstRunComplete() {
  return isCachedVersionInstalled(PACKAGE_VERSION) && fs.existsSync(setupFlagPath);
}

function markFirstRunComplete() {
  try {
    fs.mkdirSync(path.dirname(setupFlagPath), { recursive: true });
    fs.writeFileSync(setupFlagPath, "");
  } catch {}
}

async function launchTUI(binaryPath, startScreen = "launch") {
  let toolCount = 40;
  let version = PACKAGE_VERSION;
  try {
    const result = spawnSync(binaryPath, ["status"], {
      encoding: "utf8",
      timeout: 5000,
    });
    if (result.stdout) {
      const toolMatch = result.stdout.match(/(\d+)\s*(?:registered|tools)/i);
      if (toolMatch) toolCount = parseInt(toolMatch[1], 10);
      const versionMatch = result.stdout.match(/v(\d+\.\d+\.\d+)/i);
      if (versionMatch) version = versionMatch[1];
    }
  } catch {}
  const { startTUI } = await import("../tui/index.js");
  const { waitUntilExit } = startTUI({
    version,
    toolCount,
    binaryPath,
    startScreen,
    onStartServer: (mode) => {
      const serverArgs = mode === "start-http" ? ["start", "--http"] : ["start"];
      const result = spawnSync(binaryPath, serverArgs, { stdio: "inherit" });
      process.exit(result.status || 0);
    },
  });
  await waitUntilExit();
}

async function main() {
  const args = process.argv.slice(2);
  let binaryPath = resolveBinaryPath();

  if (!binaryPath || binaryPath !== getCacheBinaryPath()) {
    try {
      binaryPath = await install({ version: PACKAGE_VERSION });
    } catch (error) {
      const fallback = getFallbackBinaryPath();
      if (!fileExists(fallback)) {
        console.error(error.message);
        process.exit(1);
      }
      console.error(`${error.message}\nStarting bundled binary instead...`);
      binaryPath = fallback;
    }
  }

  if (shouldLaunchTUI(args)) {
    try {
      const startScreen = isFirstRunComplete() ? "launch" : "setup";
      await launchTUI(binaryPath, startScreen);
      if (startScreen === "setup") {
        markFirstRunComplete();
      }
      return;
    } catch (err) {
      console.error(`TUI unavailable: ${err.message}`);
      console.error("Falling back to direct mode.\n");
    }
  }

  const binaryArgs = args[0] === "tui" ? args.slice(1) : args;
  const result = spawnSync(binaryPath, binaryArgs, { stdio: "inherit" });
  if (typeof result.status === "number") {
    process.exit(result.status);
  }
  if (result.error) {
    console.error(`Could not start Relay from ${binaryPath}.\n${result.error.message}`);
    process.exit(1);
  }
  process.exit(1);
}

main().catch((error) => {
  console.error(`Relay launcher failed.\n${error.message}`);
  process.exit(1);
});

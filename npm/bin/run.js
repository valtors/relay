#!/usr/bin/env node

const { spawnSync } = require("child_process");
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

async function main() {
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

  const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: "inherit" });

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

#!/usr/bin/env node

const crypto = require("crypto");
const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");
const {
  BIN_NAME,
  PACKAGE_VERSION,
  REPO,
  downloadJson,
  downloadText,
  downloadToFile,
  ensureDir,
  fileExists,
  getCacheBinDir,
  getCacheBinaryPath,
  getFallbackBinaryPath,
  getPlatform,
  getReleaseBaseUrl,
  isCachedVersionInstalled,
  writeInstalledVersion,
} = require("./lib");

function parseArgs(argv) {
  const parsed = {
    force: false,
    version: process.env.RELAY_VERSION || PACKAGE_VERSION,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];

    if (arg === "--force") {
      parsed.force = true;
      continue;
    }

    if (arg === "--version") {
      parsed.version = argv[index + 1] || parsed.version;
      index += 1;
    }
  }

  return parsed;
}

function assetName(version) {
  const { os, arch } = getPlatform();
  const extension = os === "windows" ? "zip" : "tar.gz";
  return {
    os,
    arch,
    extension,
    fileName: `relay_${version}_${os}_${arch}.${extension}`,
  };
}

function parseChecksums(checksums) {
  const entries = new Map();

  for (const line of checksums.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed) {
      continue;
    }

    const match = trimmed.match(/^([a-fA-F0-9]{64})\s+\*?(.+)$/);
    if (match) {
      entries.set(match[2], match[1].toLowerCase());
    }
  }

  return entries;
}

function sha256(filePath) {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash("sha256");
    const stream = fs.createReadStream(filePath);
    stream.on("error", reject);
    stream.on("data", (chunk) => hash.update(chunk));
    stream.on("end", () => resolve(hash.digest("hex")));
  });
}

function runCommand(command, args) {
  const result = spawnSync(command, args, { stdio: "pipe" });
  if (!result.error && result.status === 0) {
    return;
  }

  const stderr = result.stderr ? result.stderr.toString().trim() : "";
  const stdout = result.stdout ? result.stdout.toString().trim() : "";
  const details = [stderr, stdout].filter(Boolean).join("\n");
  throw new Error(
    result.error
      ? `${command} is required to extract Relay archives but is not available.`
      : `${command} failed while extracting Relay.${details ? `\n${details}` : ""}`
  );
}

function isNotFoundError(error) {
  return /Request failed \(404\):/.test(error.message);
}

async function getLatestReleaseVersion() {
  const release = await downloadJson(`https://api.github.com/repos/${REPO}/releases/latest`);
  const version = `${release.tag_name || ""}`.replace(/^v/, "");

  if (!version) {
    throw new Error("GitHub did not return a latest Relay release tag.");
  }

  return version;
}

function expandArchive(archivePath, destinationDir, extension) {
  ensureDir(destinationDir);

  if (extension === "zip") {
    try {
      runCommand("tar", ["-xf", archivePath, "-C", destinationDir, BIN_NAME]);
      return;
    } catch (error) {
      if (process.platform !== "win32") {
        throw error;
      }
    }

    const psCommand = [
      "$ErrorActionPreference = 'Stop'",
      `Expand-Archive -LiteralPath '${archivePath.replace(/'/g, "''")}' -DestinationPath '${destinationDir.replace(/'/g, "''")}' -Force`,
    ].join("; ");
    runCommand("powershell.exe", ["-NoProfile", "-Command", psCommand]);
    return;
  }

  runCommand("tar", ["-xzf", archivePath, "-C", destinationDir, BIN_NAME]);
}

async function installVersion(version, cachedVersionLabel = version) {
  const releaseBaseUrl = getReleaseBaseUrl(version);
  const archive = assetName(version);
  const { os, arch, extension, fileName } = archive;
  const checksumUrl = `${releaseBaseUrl}/checksums.txt`;
  const archiveUrl = `${releaseBaseUrl}/${fileName}`;
  const cacheDir = getCacheBinDir();
  const archivePath = path.join(cacheDir, fileName);
  const binaryPath = getCacheBinaryPath();
  const fallbackPath = getFallbackBinaryPath();

  console.error(`Downloading Relay v${version} for ${os}/${arch}...`);

  try {
    ensureDir(cacheDir);

    const checksumsText = await downloadText(checksumUrl);
    await downloadToFile(archiveUrl, archivePath);

    const expectedChecksum = parseChecksums(checksumsText).get(fileName);
    if (!expectedChecksum) {
      throw new Error(`checksums.txt does not include ${fileName}.`);
    }

    const actualChecksum = await sha256(archivePath);
    if (actualChecksum !== expectedChecksum) {
      throw new Error(
        `Checksum mismatch for ${fileName}. Expected ${expectedChecksum}, got ${actualChecksum}.`
      );
    }

    fs.rmSync(binaryPath, { force: true });
    expandArchive(archivePath, cacheDir, extension);

    if (!fileExists(binaryPath)) {
      throw new Error(`Archive ${fileName} did not produce ${BIN_NAME}.`);
    }

    if (process.platform !== "win32") {
      fs.chmodSync(binaryPath, 0o755);
    }

    writeInstalledVersion(cachedVersionLabel);
    console.error(`Relay v${version} installed to ${binaryPath}`);
    return binaryPath;
  } catch (error) {
    const releaseUrl = `https://github.com/${REPO}/releases/tag/v${version}`;
    const availablePath = fileExists(fallbackPath) ? ` Falling back to packaged binary at ${fallbackPath}.` : "";
    throw new Error(
      `Failed to install Relay v${version} for ${os}/${arch}.\n${error.message}\nDownload manually from ${releaseUrl}.${availablePath}`
    );
  } finally {
    fs.rmSync(archivePath, { force: true });
  }
}

async function install(options = {}) {
  const requestedVersion = options.version || PACKAGE_VERSION;
  if (!options.force && isCachedVersionInstalled(requestedVersion)) {
    return getCacheBinaryPath();
  }

  try {
    return await installVersion(requestedVersion, requestedVersion);
  } catch (error) {
    if (!isNotFoundError(error)) {
      throw error;
    }

    let latestVersion;
    try {
      latestVersion = await getLatestReleaseVersion();
    } catch {
      throw error;
    }

    if (latestVersion === requestedVersion) {
      throw error;
    }

    return installVersion(latestVersion, requestedVersion);
  }
}

module.exports = {
  install,
};

if (require.main === module) {
  install(parseArgs(process.argv.slice(2))).catch((error) => {
    console.error(error.message);
    process.exit(1);
  });
}

#!/usr/bin/env node

// Downloads the correct relay binary for the current platform on npm install

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const https = require("https");

const REPO = "valtors/relay";
const BIN_DIR = path.join(__dirname);
const BIN_NAME = process.platform === "win32" ? "relay.exe" : "relay";
const BIN_PATH = path.join(BIN_DIR, BIN_NAME);

function getPlatform() {
  const os = process.platform;
  const arch = process.arch;

  const osMap = { darwin: "darwin", linux: "linux", win32: "windows" };
  const archMap = { x64: "amd64", arm64: "arm64" };

  const goOS = osMap[os];
  const goArch = archMap[arch];

  if (!goOS || !goArch) {
    console.error(`unsupported platform: ${os}/${arch}`);
    process.exit(1);
  }

  return { os: goOS, arch: goArch };
}

function getLatestVersion() {
  return new Promise((resolve, reject) => {
    https.get(
      `https://api.github.com/repos/${REPO}/releases/latest`,
      { headers: { "User-Agent": "relay-npm" } },
      (res) => {
        let data = "";
        res.on("data", (chunk) => (data += chunk));
        res.on("end", () => {
          try {
            const json = JSON.parse(data);
            resolve(json.tag_name.replace(/^v/, ""));
          } catch (e) {
            reject(new Error("could not parse release info"));
          }
        });
      }
    ).on("error", reject);
  });
}

async function install() {
  if (fs.existsSync(BIN_PATH)) {
    console.log("relay binary already exists, skipping download");
    return;
  }

  const { os, arch } = getPlatform();

  let version;
  try {
    version = await getLatestVersion();
  } catch (e) {
    console.error("could not fetch latest release. install manually from:");
    console.error(`https://github.com/${REPO}/releases`);
    process.exit(1);
  }

  const ext = os === "windows" ? "zip" : "tar.gz";
  const fileName = `relay_${version}_${os}_${arch}.${ext}`;
  const url = `https://github.com/${REPO}/releases/download/v${version}/${fileName}`;

  console.log(`downloading relay v${version} for ${os}/${arch}...`);

  const tmpFile = path.join(BIN_DIR, fileName);

  try {
    execSync(`curl -fsSL "${url}" -o "${tmpFile}"`, { stdio: "inherit" });

    if (ext === "tar.gz") {
      execSync(`tar -xzf "${tmpFile}" -C "${BIN_DIR}" relay`, { stdio: "inherit" });
    } else {
      execSync(`tar -xf "${tmpFile}" -C "${BIN_DIR}" relay.exe`, { stdio: "inherit" });
    }

    fs.unlinkSync(tmpFile);
    if (os !== "windows") {
      fs.chmodSync(BIN_PATH, 0o755);
    }

    console.log(`relay v${version} installed`);
  } catch (e) {
    console.error("download failed. install manually from:");
    console.error(`https://github.com/${REPO}/releases`);
    if (fs.existsSync(tmpFile)) fs.unlinkSync(tmpFile);
    process.exit(1);
  }
}

install();

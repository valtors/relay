const fs = require("fs");
const os = require("os");
const path = require("path");
const https = require("https");
const packageJson = require("../package.json");
const REPO = "valtors/relay";
const PACKAGE_VERSION = packageJson.version;
const BIN_NAME = process.platform === "win32" ? "relay.exe" : "relay";
function getPlatform() {
  const osMap = { darwin: "darwin", linux: "linux", win32: "windows" };
  const archMap = { x64: "amd64", arm64: "arm64" };
  const goOS = osMap[process.platform];
  const goArch = archMap[process.arch];
  if (!goOS || !goArch) {
    throw new Error(`Unsupported platform: ${process.platform}/${process.arch}.`);
  }
  return { os: goOS, arch: goArch };
}
function getCacheRoot() {
  if (process.platform === "win32") {
    return path.join(
      process.env.LOCALAPPDATA || path.join(os.homedir(), "AppData", "Local"),
      "Relay"
    );
  }
  if (process.platform === "darwin") {
    return path.join(os.homedir(), "Library", "Caches", "relay");
  }
  return path.join(process.env.XDG_CACHE_HOME || path.join(os.homedir(), ".cache"), "relay");
}
function getCacheBinDir() {
  return path.join(getCacheRoot(), "bin");
}
function getCacheBinaryPath() {
  return path.join(getCacheBinDir(), BIN_NAME);
}
function getCacheVersionPath() {
  return path.join(getCacheRoot(), "version.txt");
}
function getFallbackBinaryPath() {
  return path.join(__dirname, BIN_NAME);
}
function getBinaryCandidates() {
  return [getCacheBinaryPath(), getFallbackBinaryPath()];
}
function ensureDir(dirPath) {
  fs.mkdirSync(dirPath, { recursive: true });
}
function fileExists(filePath) {
  try {
    return fs.statSync(filePath).isFile();
  } catch {
    return false;
  }
}
function readInstalledVersion() {
  try {
    return fs.readFileSync(getCacheVersionPath(), "utf8").trim();
  } catch {
    return "";
  }
}
function writeInstalledVersion(version) {
  ensureDir(path.dirname(getCacheVersionPath()));
  fs.writeFileSync(getCacheVersionPath(), `${version}\n`, "utf8");
}
function isCachedVersionInstalled(version = PACKAGE_VERSION) {
  return fileExists(getCacheBinaryPath()) && readInstalledVersion() === version;
}
function getReleaseBaseUrl(version) {
  return `https://github.com/${REPO}/releases/download/v${version}`;
}
function httpRequest(url, options = {}, redirectCount = 0) {
  const headers = {
    Accept: "application/octet-stream",
    "User-Agent": `valtors-relay-npm/${PACKAGE_VERSION} (https://github.com/${REPO})`,
    ...(options.headers || {}),
  };
  const githubToken = process.env.GITHUB_TOKEN || process.env.GH_TOKEN;
  if (githubToken) {
    headers.Authorization = `Bearer ${githubToken}`;
  }
  return new Promise((resolve, reject) => {
    const req = https.get(
      url,
      { headers },
      (res) => {
        if (
          res.statusCode &&
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location
        ) {
          if (redirectCount >= 5) {
            reject(new Error(`Too many redirects while fetching ${url}.`));
            return;
          }
          res.resume();
          resolve(httpRequest(res.headers.location, options, redirectCount + 1));
          return;
        }
        if (!res.statusCode || res.statusCode < 200 || res.statusCode >= 300) {
          let body = "";
          res.setEncoding("utf8");
          res.on("data", (chunk) => (body += chunk));
          res.on("end", () => {
            reject(
              new Error(`Request failed (${res.statusCode || "unknown"}): ${url}${body ? `\n${body}` : ""}`)
            );
          });
          return;
        }
        resolve(res);
      }
    );
    req.on("error", reject);
  });
}
async function downloadToFile(url, destinationPath) {
  ensureDir(path.dirname(destinationPath));
  const response = await httpRequest(url);
  await new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destinationPath);
    response.pipe(file);
    file.on("finish", () => file.close(resolve));
    file.on("error", (error) => {
      fs.rmSync(destinationPath, { force: true });
      reject(error);
    });
    response.on("error", (error) => {
      fs.rmSync(destinationPath, { force: true });
      reject(error);
    });
  });
}
async function downloadText(url) {
  const response = await httpRequest(url);
  response.setEncoding("utf8");
  return new Promise((resolve, reject) => {
    let data = "";
    response.on("data", (chunk) => (data += chunk));
    response.on("end", () => resolve(data));
    response.on("error", reject);
  });
}
async function downloadJson(url) {
  const response = await httpRequest(
    url,
    { headers: { Accept: "application/vnd.github+json" } }
  );
  response.setEncoding("utf8");
  return new Promise((resolve, reject) => {
    let data = "";
    response.on("data", (chunk) => (data += chunk));
    response.on("end", () => {
      try {
        resolve(JSON.parse(data));
      } catch (error) {
        reject(new Error(`Invalid JSON response from ${url}.`));
      }
    });
    response.on("error", reject);
  });
}
module.exports = {
  BIN_NAME,
  PACKAGE_VERSION,
  REPO,
  downloadJson,
  downloadText,
  downloadToFile,
  ensureDir,
  fileExists,
  getBinaryCandidates,
  getCacheBinDir,
  getCacheBinaryPath,
  getCacheRoot,
  getCacheVersionPath,
  getFallbackBinaryPath,
  getPlatform,
  getReleaseBaseUrl,
  isCachedVersionInstalled,
  readInstalledVersion,
  writeInstalledVersion,
};

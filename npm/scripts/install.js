const fs = require("fs");
const path = require("path");
const https = require("https");

const VERSION = require("../package.json").version;
const BIN_DIR = path.join(__dirname, "..", "bin");

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

function getBinaryName() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];
  if (!platform || !arch) {
    throw new Error(
      `Unsupported platform: ${process.platform}-${process.arch}`
    );
  }
  const ext = process.platform === "win32" ? ".exe" : "";
  return `remmd-${platform}-${arch}${ext}`;
}

function download(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return download(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
        }
        const chunks = [];
        res.on("data", (c) => chunks.push(c));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

async function main() {
  const binaryName = getBinaryName();
  const url = `https://github.com/lagz0ne/remmd/releases/download/v${VERSION}/${binaryName}`;
  const dest = path.join(BIN_DIR, process.platform === "win32" ? "remmd-bin.exe" : "remmd-bin");

  console.log(`Downloading remmd v${VERSION} for ${process.platform}-${process.arch}...`);

  const data = await download(url);
  fs.mkdirSync(BIN_DIR, { recursive: true });
  fs.writeFileSync(dest, data);
  fs.chmodSync(dest, 0o755);

  console.log(`Installed remmd to ${dest}`);
}

main().catch((err) => {
  console.error(`Failed to install remmd: ${err.message}`);
  process.exit(1);
});

// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Postinstall de huginn-tui: descarga el binario del release de GitHub.
// Solo usa stdlib de Node (https/fs/os/path/child_process), sin dependencias.
// Sirve igual para npm, pnpm y bun (todos ejecutan `postinstall`).
//
// PRERREQUISITO: debe existir el release v0.2.0 en
// https://github.com/gpb-codes/huginn-tui/releases con el asset de tu
// plataforma. Si no existe, FALLA con mensaje claro (prohibido el éxito falso).
"use strict";

const fs = require("fs");
const https = require("https");
const os = require("os");
const path = require("path");
const { execFileSync } = require("child_process");

// Versión pineada (ver palette.go: const VERSION = "v0.2.0").
const VERSION = "v0.2.0";
const REPO = "gpb-codes/huginn-tui";

// Mapa plataforma Node -> asset del release.
// Combos soportados: win32-x64, darwin-arm64, darwin-x64, linux-x64, linux-arm64.
function assetPara(plataforma, arquitectura) {
  const clave = `${plataforma}-${arquitectura}`;
  const mapa = {
    "win32-x64": "huginn-windows-amd64.exe",
    "darwin-arm64": "huginn-darwin-arm64.tar.gz",
    "darwin-x64": "huginn-darwin-amd64.tar.gz",
    "linux-x64": "huginn-linux-amd64.tar.gz",
    "linux-arm64": "huginn-linux-arm64.tar.gz",
  };
  return { clave, asset: mapa[clave] || null };
}

// Descarga siguiendo redirecciones (GitHub redirige a objects.githubusercontent.com).
function descargar(url, destino, redirecciones = 5) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        const codigo = res.statusCode || 0;
        if (codigo >= 300 && codigo < 400 && res.headers.location) {
          res.resume();
          if (redirecciones <= 0) {
            reject(new Error("demasiadas redirecciones al descargar el release"));
            return;
          }
          descargar(res.headers.location, destino, redirecciones - 1).then(resolve, reject);
          return;
        }
        if (codigo === 404) {
          res.resume();
          reject(new Error("HTTP 404: el release o el asset no existe"));
          return;
        }
        if (codigo !== 200) {
          res.resume();
          reject(new Error(`HTTP ${codigo} al descargar el release`));
          return;
        }
        const fichero = fs.createWriteStream(destino);
        res.pipe(fichero);
        fichero.on("finish", () => fichero.close(resolve));
        fichero.on("error", (err) => {
          fs.rmSync(destino, { force: true });
          reject(err);
        });
      })
      .on("error", reject);
  });
}

async function main() {
  const { clave, asset } = assetPara(process.platform, process.arch);
  if (!asset) {
    console.error(
      `huginn: ERROR: plataforma no soportada: ${clave} ` +
        `(soportadas: win32-x64, darwin-arm64, darwin-x64, linux-x64, linux-arm64)`
    );
    process.exit(1);
  }

  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${asset}`;
  const dirBin = path.join(__dirname, "bin");
  const esWindows = asset.endsWith(".exe");
  const nombreBinario = esWindows ? "huginn.exe" : "huginn";
  const destinoBinario = path.join(dirBin, nombreBinario);

  // Si el binario ya está instalado (reinstalación), no descarga de nuevo.
  if (fs.existsSync(destinoBinario)) {
    console.log(`huginn: binario ya presente en ${destinoBinario}, omito descarga`);
    return;
  }

  console.log(`huginn: descargando ${url} ...`);
  const tmp = path.join(os.tmpdir(), asset);
  try {
    await descargar(url, tmp);
  } catch (err) {
    console.error(`huginn: ERROR: no se pudo descargar ${asset}: ${err.message}`);
    console.error(
      `huginn: prerrequisito: crea el release v${VERSION} en https://github.com/${REPO}/releases ` +
        `y sube el asset ${asset} (o compila manual: ver docs/install.md, Go 1.25+)`
    );
    process.exit(1);
  }

  try {
    if (esWindows) {
      // En Windows el asset ya es el .exe: lo coloca junto al lanzador.
      fs.mkdirSync(dirBin, { recursive: true });
      fs.copyFileSync(tmp, destinoBinario);
    } else {
      // En Unix el asset es un .tar.gz con un único `huginn` dentro.
      const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "huginn-"));
      try {
        execFileSync("tar", ["-xzf", tmp, "-C", tmpDir], { stdio: "inherit" });
        const extraido = path.join(tmpDir, "huginn");
        if (!fs.existsSync(extraido)) {
          console.error(`huginn: ERROR: el tarball ${asset} no contiene el binario 'huginn'`);
          process.exit(1);
        }
        fs.mkdirSync(dirBin, { recursive: true });
        fs.copyFileSync(extraido, destinoBinario);
        fs.chmodSync(destinoBinario, 0o755);
      } finally {
        fs.rmSync(tmpDir, { recursive: true, force: true });
      }
    }
    console.log(`huginn: instalado en ${destinoBinario} (v${VERSION} ${clave})`);
  } finally {
    fs.rmSync(tmp, { force: true });
  }
}

main().catch((err) => {
  console.error(`huginn: ERROR inesperado en postinstall: ${err.message}`);
  process.exit(1);
});

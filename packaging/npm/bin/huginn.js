#!/usr/bin/env node
// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Lanzador del binario `huginn` descargado por postinstall (install.js).
// Localiza el binario junto a este fichero, lo ejecuta con los mismos
// argumentos y propaga su exit code / señal (nunca finge éxito).
"use strict";

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

// El binario vive junto a este lanzador: huginn.exe en Windows, huginn en Unix.
const binario = path.join(__dirname, process.platform === "win32" ? "huginn.exe" : "huginn");

if (!fs.existsSync(binario)) {
  console.error(
    `huginn: ERROR: no se encontró el binario en ${binario}. ` +
      `El postinstall no se ejecutó o falló (prerrequisito: release v0.2.0 con tu asset en ` +
      `https://github.com/gpb-codes/huginn-tui/releases). Reinstala con tu gestor favorito.`
  );
  process.exit(1);
}

const resultado = spawnSync(binario, process.argv.slice(2), { stdio: "inherit" });
if (resultado.error) {
  console.error(`huginn: ERROR al ejecutar el binario: ${resultado.error.message}`);
  process.exit(1);
}
// Propaga la señal o el exit code del binario (éxito solo si el binario tuvo éxito).
if (resultado.signal) {
  process.kill(process.pid, resultado.signal);
} else {
  process.exit(resultado.status === null ? 1 : resultado.status);
}

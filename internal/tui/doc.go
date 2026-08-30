package tui

// Package tui contendrá el runtime de Huginn (model, Update, View).
// Actualmente vive en root main.go:460 (model) y main.go:1587 (View).
// Migración en curso: mover model/View/viewServers/viewSettings/viewChat aquí
// para que cmd/huginn/main.go pueda hacer tui.Run(ctx) sin depender de main.
//
// El TUI depende de application/ports, nunca directamente de infrastructure/vault.

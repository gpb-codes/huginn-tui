// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Binario huginn-bot: control remoto LOCAL para Huginn (solo stdlib).
//
// Escucha en loopback por defecto y exige HUGINN_BOT_TOKEN salvo modo
// inseguro loopback explícito. Reutiliza bootstrap.New como la TUI.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"huginn/internal/bootstrap"
	"huginn/internal/bot"
)

func main() {
	// Flags de escucha y contexto (defecto: solo loopback local).
	addr := flag.String("addr", "127.0.0.1:8765", "dirección de escucha host:puerto")
	vault := flag.String("vault", "", "ruta del vault (Agent Vault)")
	project := flag.String("project", ".", "ruta del proyecto activo")
	insecure := flag.Bool("allow-insecure-loopback", false, "permite arrancar sin token solo en loopback")
	flag.Parse()

	// El token solo viaja por entorno; jamás por flags ni logs.
	token := os.Getenv("HUGINN_BOT_TOKEN")

	// Compone el App igual que la TUI (Registry, providers, traza).
	app, err := bootstrap.New(*project, *vault)
	if err != nil {
		fmt.Fprintf(os.Stderr, "huginn-bot: no se pudo construir la app: %v\n", err)
		os.Exit(1)
	}

	// Crea el servidor (valida token/loopback antes de escuchar).
	srv, err := bot.New(bot.Config{
		Addr:                  *addr,
		Token:                 token,
		VaultPath:             *vault,
		ProjectPath:           *project,
		AllowInsecureLoopback: *insecure,
	}, app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "huginn-bot: %v\n", err)
		os.Exit(1)
	}

	// Apagado graceful ante SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("huginn-bot escuchando en http://%s\n", srv.Addr())
	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "huginn-bot: %v\n", err)
		os.Exit(1)
	}
}

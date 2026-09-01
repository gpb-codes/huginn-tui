package tui

import tea "charm.land/bubbletea/v2"

// App is the TUI application entry. Future: model will be moved here from main.go.
type App struct {
	program *tea.Program
}

func New(program *tea.Program) *App { return &App{program: program} }

func (a *App) Run() (tea.Model, error) { return a.program.Run() }

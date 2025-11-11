package main

import (
    "log"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/anakafeel/LazyDots/internal/config"
    "github.com/anakafeel/LazyDots/internal/tui"
)

func main() {
    // If config doesn’t exist → first time setup
    if !config.Exists() {
        runSetup()
        return
    }

    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Launch main TUI (with config)
    app := tui.New(cfg)
    if _, err := tea.NewProgram(app).Run(); err != nil {
        log.Fatal(err)
    }
}

func runSetup() {
    p := tea.NewProgram(tui.NewSetupModel())
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}

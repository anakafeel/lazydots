package main

import (
	"log"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anakafeel/LazyDots/internal/config"
	"github.com/anakafeel/LazyDots/internal/tui"
)

func main() {
	// If config doesn't exist → first time setup (no splash)
	if !config.Exists() {
		runSetup()
		return
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Pick banner color once for consistent branding
	bannerColor := tui.PickBannerColor()

	// Check for --no-splash flag
	skipSplash := slices.Contains(os.Args[1:], "--no-splash")

	// Choose initial model
	var initialModel tea.Model
	if skipSplash {
		initialModel = tui.New(cfg, bannerColor, 0, 0)
	} else {
		initialModel = tui.NewSplashModel(cfg, bannerColor)
	}

	// Launch TUI
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func runSetup() {
	p := tea.NewProgram(tui.NewSetupModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

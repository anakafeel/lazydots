package tui

import (
	"fmt"

	"github.com/anakafeel/LazyDots/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cfg           config.Config
	width, height int
}

func New(cfg config.Config) model {
	return model{cfg: cfg}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Save window dimensions so child views (lists) can size themselves.
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			// Go to setup screen (it tells you to restart when done)
			return NewSetupModel(), nil

		case "l":
			// Go to the package list view (Stow-style packages)
			return NewPackageListModel(m.cfg.DotfilesPath, m.width, m.height), nil
		}
	}

	return m, nil
}

func (m model) View() string {
	header := "🧩 Welcome to LazyDots — Manage your dotfiles like a pro\n"
	divider := "────────────────────────────────────────────\n"
	currentPath := fmt.Sprintf("📁 Current Dotfiles Path:\n%s\n\n", m.cfg.DotfilesPath)

	help := `🧠 Quick Actions:
  [r]  → Reconfigure path
  [l]  → List dotfile packages
  [q]  → Quit

💡 Tip: Point LazyDots at your Stow-style dotfiles repo (e.g. ~/linuxworkspace/dotfiles).
`

	footer := "\n────────────────────────────────────────────\n"

	return header + divider + currentPath + help + footer
}

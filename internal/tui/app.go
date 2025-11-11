package tui

import (
	"fmt"

	"github.com/anakafeel/LazyDots/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// Message for switching between models
type switchModelMsg struct {
	model tea.Model
}

// Helper command to trigger model switch
func switchTo(newModel tea.Model) tea.Cmd {
	return func() tea.Msg {
		return switchModelMsg{model: newModel}
	}
}

type model struct {
	cfg config.Config
	msg string
}

func New(cfg config.Config) model {
	return model{cfg: cfg}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, switchTo(NewSetupModel())
		case "l":
    		return m, switchTo(NewFileListModel(m.cfg.DotfilesPath))
		}
	case switchModelMsg:
		return msg.model, nil
	}
	return m, nil
}

func (m model) View() string {
	header := "🧩 Welcome to LazyDots — Manage your dotfiles like a pro\n"
	divider := "────────────────────────────────────────────\n"
	currentPath := fmt.Sprintf("📁 Current Dotfiles Path:\n%s\n\n", m.cfg.DotfilesPath)

	help := `🧠 Quick Actions:
  [r]  → Reconfigure path
  [l]  → List dotfiles
  [q]  → Quit

💡 Tip: Keep your environment portable, sync dotfiles across devices effortlessly.
`

	footer := "\n────────────────────────────────────────────\n"

	return header + divider + currentPath + help + footer
}

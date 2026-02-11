package tui

import (
	"time"

	"github.com/anakafeel/LazyDots/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const splashDuration = 1500 * time.Millisecond

type splashModel struct {
	cfg           config.Config
	bannerColor   string
	width, height int
}

type splashDoneMsg struct{}

func NewSplashModel(cfg config.Config, bannerColor string) splashModel {
	return splashModel{cfg: cfg, bannerColor: bannerColor}
}

func (m splashModel) Init() tea.Cmd {
	return tea.Tick(splashDuration, func(t time.Time) tea.Msg {
		return splashDoneMsg{}
	})
}

func (m splashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
		// Any other key → transition to dashboard
		dashboard := New(m.cfg, m.bannerColor, m.width, m.height)
		return dashboard, nil

	case splashDoneMsg:
		dashboard := New(m.cfg, m.bannerColor, m.width, m.height)
		return dashboard, nil
	}
	return m, nil
}

func (m splashModel) View() string {
	banner := RenderBanner(m.width, m.bannerColor)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		banner,
		"",
		subtitleStyle.Render("Terminal dotfile manager"),
		"",
		hintStyle.Render("Press any key to continue (q to quit)..."),
	)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}

	return content
}

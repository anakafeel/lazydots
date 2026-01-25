package tui

import (
	"fmt"
	"path/filepath"

	"github.com/anakafeel/LazyDots/internal/config"
	"github.com/anakafeel/LazyDots/internal/git"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	cfg           config.Config
	bannerColor   string
	gitStatus     git.RepoStatus
	width, height int

	// Commit input mode
	committing  bool
	commitInput textinput.Model
	statusMsg   string

	// Help panel
	showHelp bool
}

func New(cfg config.Config, bannerColor string) model {
	ti := textinput.New()
	ti.Placeholder = "Enter commit message..."
	ti.CharLimit = 200

	return model{
		cfg:         cfg,
		bannerColor: bannerColor,
		gitStatus:   git.GetStatus(cfg.DotfilesPath),
		commitInput: ti,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle commit input mode
		if m.committing {
			switch msg.String() {
			case "enter":
				message := m.commitInput.Value()
				if message == "" {
					m.statusMsg = "⚠️ Commit message cannot be empty"
					m.committing = false
					m.commitInput.Reset()
					return m, nil
				}
				if err := git.Commit(m.cfg.DotfilesPath, message); err != nil {
					m.statusMsg = "⚠️ " + err.Error()
				} else {
					m.statusMsg = "✅ Committed: " + message
				}
				m.committing = false
				m.commitInput.Reset()
				m.commitInput.Blur()
				m.gitStatus = git.GetStatus(m.cfg.DotfilesPath)
				return m, nil

			case "esc":
				m.committing = false
				m.commitInput.Reset()
				m.commitInput.Blur()
				m.statusMsg = ""
				return m, nil
			}

			var cmd tea.Cmd
			m.commitInput, cmd = m.commitInput.Update(msg)
			return m, cmd
		}

		// Normal mode
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			return NewSetupModel(), nil

		case "l":
			return NewPackageListModel(m.cfg.DotfilesPath, m.bannerColor, m.width, m.height), nil

		case "c":
			m.committing = true
			m.commitInput.Focus()
			m.statusMsg = ""
			return m, textinput.Blink

		case "p":
			if err := git.Push(m.cfg.DotfilesPath); err != nil {
				m.statusMsg = "⚠️ " + err.Error()
			} else {
				m.statusMsg = "✅ Pushed to remote"
			}
			m.gitStatus = git.GetStatus(m.cfg.DotfilesPath)
			return m, nil

		case "P":
			if err := git.Pull(m.cfg.DotfilesPath); err != nil {
				m.statusMsg = "⚠️ " + err.Error()
			} else {
				m.statusMsg = "✅ Pulled from remote"
			}
			m.gitStatus = git.GetStatus(m.cfg.DotfilesPath)
			return m, nil

		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	banner := RenderBanner(m.width, m.bannerColor)

	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("────────────────────────────────────────────────────────────────────")

	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	gitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	// Git status line
	repoName := filepath.Base(m.cfg.DotfilesPath)
	gitLine := fmt.Sprintf("📦 %s %s", repoName, gitStyle.Render(m.gitStatus.FormatStatus()))

	currentPath := fmt.Sprintf("%s\n%s",
		labelStyle.Render("📁 Path:"),
		pathStyle.Render(m.cfg.DotfilesPath))

	help := "\n💡 Press [?] for help"

	// Full help panel
	if m.showHelp {
		helpStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

		helpContent := `Navigation
  ↑/↓      navigate list
  enter    select package
  esc/q    back / quit

File Operations
  space    toggle link/unlink
  a        link all files
  A        unlink all files

Git
  c        commit changes
  p        push to remote
  P        pull from remote

General
  l        list packages
  r        reconfigure path
  ?        toggle help
  q        quit`

		help = "\n" + helpStyle.Render(helpContent)
	}

	// Show commit input if in commit mode
	var inputSection string
	if m.committing {
		inputSection = fmt.Sprintf("\n💬 Commit message (esc to cancel):\n%s\n", m.commitInput.View())
	}

	// Show status message if any
	var statusSection string
	if m.statusMsg != "" {
		statusSection = "\n" + m.statusMsg
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		banner,
		"",
		divider,
		"",
		gitLine,
		currentPath,
		inputSection,
		statusSection,
		help,
		"",
		divider,
	)
}

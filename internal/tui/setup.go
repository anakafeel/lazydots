package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anakafeel/LazyDots/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type setupModel struct {
	input textinput.Model
	msg   string
	done  bool
	err   error
}

func NewSetupModel() setupModel {
	ti := textinput.New()
	ti.Placeholder = "/home/user/dotfiles"
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 50
	return setupModel{input: ti}
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			path := m.input.Value()

			absPath, err := filepath.Abs(path)
			if err != nil {
				m.msg = fmt.Sprintf("❌ Invalid path: %v", err)
				return m, nil
			}

			info, err := os.Stat(absPath)
			if os.IsNotExist(err) {
				m.msg = "❌ That directory doesn’t exist."
				return m, nil
			}
			if err != nil {
				m.msg = fmt.Sprintf("❌ Error checking path: %v", err)
				return m, nil
			}
			if !info.IsDir() {
				m.msg = "❌ That path is not a directory."
				return m, nil
			}

			// --- Optional: check for dotfiles ---
			hasDotfiles := false
			entries, _ := os.ReadDir(absPath)
			for _, e := range entries {
				if len(e.Name()) > 0 && e.Name()[0] == '.' {
					hasDotfiles = true
					break
				}
			}

			if !hasDotfiles {
				m.msg = "⚠️ No dotfiles found here, but saved anyway."
			} else {
				m.msg = "✅ Valid directory! Config saved."
			}

			// --- Save config ---
			cfg := config.Config{DotfilesPath: absPath}
			m.err = config.Save(cfg)
			if m.err != nil {
				m.msg = fmt.Sprintf("❌ Failed to save config: %v", m.err)
				return m, nil
			}

			m.done = true
			return m, tea.Quit

		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m setupModel) View() string {
	if m.done {
		return "✅ Config saved! Restart LazyDots.\n"
	}

	return fmt.Sprintf(
		"🧩 Welcome to LazyDots!\n\nEnter the full path to your dotfiles directory:\n\n%s\n\n%s\n\n(press Enter to save, Esc to quit)",
		m.input.View(),
		m.msg,
	)
}

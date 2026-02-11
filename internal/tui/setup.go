package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/anakafeel/LazyDots/internal/config"
	"github.com/anakafeel/LazyDots/internal/fs"
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
	ti.Placeholder = "~/linuxworkspace/dotfiles"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60
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
			raw := m.input.Value()

			// Resolve and validate the path
			absPath, err := fs.ResolvePath(raw)
			if err != nil {
				switch {
				case errors.Is(err, fs.ErrEmptyPath):
					m.msg = "Please enter a path."
				case errors.Is(err, fs.ErrHomeExpand):
					m.msg = fmt.Sprintf("Failed to expand ~: %v", err)
				default:
					m.msg = fmt.Sprintf("Invalid path: %v", err)
				}
				return m, nil
			}

			// Validate it's an existing directory
			if err := fs.ValidateDirectory(absPath); err != nil {
				switch {
				case errors.Is(err, fs.ErrNotExist):
					m.msg = fmt.Sprintf("Directory doesn't exist: %s", absPath)
				case errors.Is(err, fs.ErrNotDirectory):
					m.msg = fmt.Sprintf("Not a directory: %s", absPath)
				default:
					m.msg = fmt.Sprintf("Error checking path: %v", err)
				}
				return m, nil
			}

			// Optional: check for dotfiles/packages
			hasEntries := false
			entries, _ := os.ReadDir(absPath)
			for _, e := range entries {
				if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
					hasEntries = true
					break
				}
			}
			if !hasEntries {
				m.msg = "Directory is empty, but config saved anyway."
			} else {
				m.msg = "Valid directory! Config saved."
			}

			cfg := config.Config{DotfilesPath: absPath}
			m.err = config.Save(cfg)
			if m.err != nil {
				m.msg = fmt.Sprintf("Failed to save config: %v", m.err)
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

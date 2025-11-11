package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anakafeel/LazyDots/internal/config"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type fileItem struct{ name string }

func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return "" }
func (f fileItem) FilterValue() string { return f.name }

type fileListModel struct {
	list list.Model
}

func NewFileListModel(path string) fileListModel {
	items := []list.Item{}

	entries, err := os.ReadDir(path)
	if err != nil {
		items = append(items, fileItem{name: fmt.Sprintf("❌ Failed to read dir: %v", err)})
	} else {
		for _, e := range entries {
			if e.Name()[0] == '.' { // only show dotfiles
				items = append(items, fileItem{name: e.Name()})
			}
		}
		if len(items) == 0 {
			items = append(items, fileItem{name: "⚠️ No dotfiles found in this directory"})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("Dotfiles in %s", filepath.Base(path))
	return fileListModel{list: l}
}

func (m fileListModel) Init() tea.Cmd { return nil }

func (m fileListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Return to main screen
			return New(config.Config{}), nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m fileListModel) View() string {
	return m.list.View()
}

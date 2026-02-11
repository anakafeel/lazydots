package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pkgEntry struct {
	name string
	path string
}

type packagesPane struct {
	width, height int
	focused       bool
	items         []pkgEntry
	cursor        int
	offset        int
}

func newPackagesPane(rootPath string) *packagesPane {
	var items []pkgEntry
	entries, err := os.ReadDir(rootPath)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() || e.Name() == ".git" {
				continue
			}
			items = append(items, pkgEntry{
				name: e.Name(),
				path: filepath.Join(rootPath, e.Name()),
			})
		}
	}
	return &packagesPane{items: items}
}

func (p *packagesPane) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch km.String() {
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
			p.ensureVisible()
		}
	case "down", "j":
		if p.cursor < len(p.items)-1 {
			p.cursor++
			p.ensureVisible()
		}
	}
	return nil
}

func (p *packagesPane) innerHeight() int {
	h := p.height - 2
	if h < 1 {
		h = 1
	}
	return h
}

func (p *packagesPane) ensureVisible() {
	ih := p.innerHeight()
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+ih {
		p.offset = p.cursor - ih + 1
	}
}

func (p *packagesPane) View() string {
	if len(p.items) == 0 {
		dim := lipgloss.NewStyle().Foreground(colorDim)
		return renderPane(p.Title(), " "+dim.Render("No packages found"), p.width, p.height, p.focused)
	}

	cursorStyle := lipgloss.NewStyle().
		Foreground(colorCursorFg).
		Background(colorCursorBg).
		Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(colorNormal)

	ih := p.innerHeight()
	innerW := p.width - 2

	var lines []string
	for i := p.offset; i < len(p.items) && i < p.offset+ih; i++ {
		label := " " + p.items[i].name
		if i == p.cursor && p.focused {
			label = cursorStyle.Render(padOrTruncate(label, innerW))
		} else {
			label = normalStyle.Render(label)
		}
		lines = append(lines, label)
	}

	return renderPane(p.Title(), strings.Join(lines, "\n"), p.width, p.height, p.focused)
}

func (p *packagesPane) Selected() *pkgEntry {
	if len(p.items) == 0 || p.cursor >= len(p.items) {
		return nil
	}
	return &p.items[p.cursor]
}

func (p *packagesPane) SetSize(w, h int) { p.width, p.height = w, h }
func (p *packagesPane) Focus()           { p.focused = true }
func (p *packagesPane) Blur()            { p.focused = false }
func (p *packagesPane) Focused() bool    { return p.focused }
func (p *packagesPane) Title() string {
	return fmt.Sprintf("2 Packages (%d)", len(p.items))
}

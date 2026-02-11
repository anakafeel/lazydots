package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type detailPane struct {
	width, height int
	focused       bool
	viewport      viewport.Model
	title         string
}

func newDetailPane() *detailPane {
	vp := viewport.New(0, 0)
	return &detailPane{
		viewport: vp,
		title:    "5 Detail",
	}
}

func (p *detailPane) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return cmd
}

func (p *detailPane) View() string {
	return renderPane(p.title, p.viewport.View(), p.width, p.height, p.focused)
}

func (p *detailPane) SetContent(title, content string) {
	p.title = title
	p.viewport.SetContent(content)
}

func (p *detailPane) SetSize(w, h int) {
	p.width, p.height = w, h
	innerW := w - 2
	innerH := h - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}
	p.viewport.Width = innerW
	p.viewport.Height = innerH
}

func (p *detailPane) Focus()        { p.focused = true }
func (p *detailPane) Blur()         { p.focused = false }
func (p *detailPane) Focused() bool { return p.focused }
func (p *detailPane) Title() string { return p.title }

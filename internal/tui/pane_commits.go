package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type commitsPane struct {
	width, height int
	focused       bool
}

func newCommitsPane() *commitsPane {
	return &commitsPane{}
}

func (p *commitsPane) Update(tea.Msg) tea.Cmd { return nil }

func (p *commitsPane) View() string {
	dim := lipgloss.NewStyle().Foreground(colorDim)
	content := " " + dim.Render("(commit log)")
	return renderPane(p.Title(), content, p.width, p.height, p.focused)
}

func (p *commitsPane) SetSize(w, h int) { p.width, p.height = w, h }
func (p *commitsPane) Focus()           { p.focused = true }
func (p *commitsPane) Blur()            { p.focused = false }
func (p *commitsPane) Focused() bool    { return p.focused }
func (p *commitsPane) Title() string    { return "4 Commits" }

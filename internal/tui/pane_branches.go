package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type branchesPane struct {
	width, height int
	focused       bool
	branch        string
}

func newBranchesPane(branch string) *branchesPane {
	return &branchesPane{branch: branch}
}

func (p *branchesPane) Update(tea.Msg) tea.Cmd { return nil }

func (p *branchesPane) View() string {
	gs := lipgloss.NewStyle().Foreground(colorGit)
	dim := lipgloss.NewStyle().Foreground(colorDim)

	content := " " + gs.Render(p.branch) + "\n " + dim.Render("(local branches)")

	return renderPane(p.Title(), content, p.width, p.height, p.focused)
}

func (p *branchesPane) SetSize(w, h int) { p.width, p.height = w, h }
func (p *branchesPane) Focus()           { p.focused = true }
func (p *branchesPane) Blur()            { p.focused = false }
func (p *branchesPane) Focused() bool    { return p.focused }
func (p *branchesPane) Title() string    { return "3 Branches" }

package tui

import (
	"fmt"

	"github.com/anakafeel/LazyDots/internal/git"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type statusPane struct {
	width, height int
	focused       bool
	gitStatus     git.RepoStatus
	repoName      string
	repoPath      string
}

func newStatusPane(repoName, repoPath string, gs git.RepoStatus) *statusPane {
	return &statusPane{
		repoName:  repoName,
		repoPath:  repoPath,
		gitStatus: gs,
	}
}

func (p *statusPane) Update(tea.Msg) tea.Cmd { return nil }

func (p *statusPane) View() string {
	gs := lipgloss.NewStyle().Foreground(colorGit)
	dim := lipgloss.NewStyle().Foreground(colorDim)
	normal := lipgloss.NewStyle().Foreground(colorNormal)

	content := fmt.Sprintf(
		" %s %s\n %s %s",
		normal.Render(p.repoName),
		gs.Render(p.gitStatus.FormatStatus()),
		dim.Render("Path:"),
		normal.Render(p.repoPath),
	)

	return renderPane(p.Title(), content, p.width, p.height, p.focused)
}

func (p *statusPane) SetSize(w, h int) { p.width, p.height = w, h }
func (p *statusPane) Focus()           { p.focused = true }
func (p *statusPane) Blur()            { p.focused = false }
func (p *statusPane) Focused() bool    { return p.focused }
func (p *statusPane) Title() string    { return "1 Status" }

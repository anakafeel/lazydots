package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/anakafeel/LazyDots/internal/config"
	"github.com/anakafeel/LazyDots/internal/git"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	paneStatus   = 0
	panePackages = 1
	paneBranches = 2
	paneCommits  = 3
	paneDetail   = 4
	paneCount    = 5
)

type model struct {
	cfg         config.Config
	bannerColor string
	width       int
	height      int

	panes      [paneCount]Pane
	focusIndex int

	committing  bool
	commitInput textinput.Model
	statusMsg   string
}

func New(cfg config.Config, bannerColor string, width, height int) model {
	gitStatus := git.GetStatus(cfg.DotfilesPath)
	repoName := filepath.Base(cfg.DotfilesPath)

	ti := textinput.New()
	ti.Placeholder = "commit message..."
	ti.CharLimit = 200

	m := model{
		cfg:         cfg,
		bannerColor: bannerColor,
		width:       width,
		height:      height,
		commitInput: ti,
	}

	m.panes[paneStatus] = newStatusPane(repoName, cfg.DotfilesPath, gitStatus)
	m.panes[panePackages] = newPackagesPane(cfg.DotfilesPath)
	m.panes[paneBranches] = newBranchesPane(gitStatus.Branch)
	m.panes[paneCommits] = newCommitsPane()
	m.panes[paneDetail] = newDetailPane()

	// Default focus: packages pane
	m.focusIndex = panePackages
	m.panes[panePackages].Focus()

	m.syncDetail()

	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Commit input mode intercepts all keys
		if m.committing {
			switch msg.String() {
			case "enter":
				message := m.commitInput.Value()
				if message == "" {
					m.statusMsg = "Commit message cannot be empty"
					m.committing = false
					m.commitInput.Reset()
					return m, nil
				}
				if err := git.Commit(m.cfg.DotfilesPath, message); err != nil {
					m.statusMsg = err.Error()
				} else {
					m.statusMsg = "Committed: " + message
				}
				m.committing = false
				m.commitInput.Reset()
				m.commitInput.Blur()
				m.refreshGit()
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

		// Global keys
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.panes[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % paneCount
			m.panes[m.focusIndex].Focus()
			m.syncDetail()
			return m, nil
		case "shift+tab":
			m.panes[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex - 1 + paneCount) % paneCount
			m.panes[m.focusIndex].Focus()
			m.syncDetail()
			return m, nil
		case "1", "2", "3", "4", "5":
			idx := int(msg.String()[0] - '1')
			m.panes[m.focusIndex].Blur()
			m.focusIndex = idx
			m.panes[m.focusIndex].Focus()
			m.syncDetail()
			return m, nil
		case "c":
			m.committing = true
			m.commitInput.Width = m.width - 12
			m.commitInput.Focus()
			m.statusMsg = ""
			return m, textinput.Blink
		case "p":
			if err := git.Push(m.cfg.DotfilesPath); err != nil {
				m.statusMsg = err.Error()
			} else {
				m.statusMsg = "Pushed to remote"
			}
			m.refreshGit()
			return m, nil
		case "P":
			if err := git.Pull(m.cfg.DotfilesPath); err != nil {
				m.statusMsg = err.Error()
			} else {
				m.statusMsg = "Pulled from remote"
			}
			m.refreshGit()
			return m, nil
		}

		// Delegate to focused pane
		cmd := m.panes[m.focusIndex].Update(msg)
		m.syncDetail()
		return m, cmd
	}

	return m, nil
}

func (m model) refreshGit() {
	gs := git.GetStatus(m.cfg.DotfilesPath)
	if sp, ok := m.panes[paneStatus].(*statusPane); ok {
		sp.gitStatus = gs
	}
	if bp, ok := m.panes[paneBranches].(*branchesPane); ok {
		bp.branch = gs.Branch
	}
}

func (m model) syncDetail() {
	dp, ok := m.panes[paneDetail].(*detailPane)
	if !ok {
		return
	}

	switch m.focusIndex {
	case panePackages:
		pp := m.panes[panePackages].(*packagesPane)
		sel := pp.Selected()
		if sel == nil {
			dp.SetContent("5 Detail", " No package selected")
			return
		}
		dp.SetContent(
			fmt.Sprintf("5 %s", sel.name),
			m.buildFilePreview(sel.path),
		)
	case paneStatus:
		dp.SetContent("5 Overview", m.buildOverview())
	default:
		dp.SetContent("5 Detail", " Select a pane with content")
	}
}

func (m model) buildFilePreview(pkgPath string) string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}

	linked := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	missing := lipgloss.NewStyle().Foreground(colorDim)
	conflict := lipgloss.NewStyle().Foreground(colorHighlight)

	var lines []string
	_ = filepath.WalkDir(pkgPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(pkgPath, path)

		targetRel := rel
		if !strings.HasPrefix(targetRel, ".") {
			if i := strings.IndexRune(targetRel, os.PathSeparator); i != -1 {
				targetRel = targetRel[i+1:]
			}
		}
		targetPath := filepath.Join(home, targetRel)
		status := computeLinkStatus(path, targetPath)

		var icon string
		switch status {
		case StatusLinked:
			icon = linked.Render("✓")
		case StatusMissing:
			icon = missing.Render("○")
		case StatusConflict:
			icon = conflict.Render("!")
		}

		lines = append(lines, fmt.Sprintf(" %s %s", icon, rel))
		return nil
	})

	if len(lines) == 0 {
		return " No files in this package"
	}
	return strings.Join(lines, "\n")
}

func (m model) buildOverview() string {
	gs := git.GetStatus(m.cfg.DotfilesPath)
	gitStyle := lipgloss.NewStyle().Foreground(colorGit)
	dim := lipgloss.NewStyle().Foreground(colorDim)
	normal := lipgloss.NewStyle().Foreground(colorNormal)

	var lines []string
	lines = append(lines, " "+normal.Render(filepath.Base(m.cfg.DotfilesPath))+" "+gitStyle.Render(gs.FormatStatus()))
	lines = append(lines, " "+dim.Render("Path:")+" "+normal.Render(m.cfg.DotfilesPath))
	lines = append(lines, "")

	pp := m.panes[panePackages].(*packagesPane)
	lines = append(lines, " "+dim.Render(fmt.Sprintf("Packages: %d", len(pp.items))))
	for _, pkg := range pp.items {
		lines = append(lines, "   "+normal.Render(pkg.name))
	}
	return strings.Join(lines, "\n")
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	if m.width < 40 || m.height < 15 {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			"Terminal too small.\nResize to at least 40x15.",
		)
	}

	ly := ComputeLayout(m.width, m.height)

	// Apply sizes to all panes
	m.panes[paneStatus].SetSize(ly.Status.Width, ly.Status.Height)
	m.panes[panePackages].SetSize(ly.Packages.Width, ly.Packages.Height)
	m.panes[paneBranches].SetSize(ly.Branches.Width, ly.Branches.Height)
	m.panes[paneCommits].SetSize(ly.Commits.Width, ly.Commits.Height)
	m.panes[paneDetail].SetSize(ly.Detail.Width, ly.Detail.Height)

	// Left column: stack 4 panes vertically
	leftCol := lipgloss.JoinVertical(lipgloss.Left,
		m.panes[paneStatus].View(),
		m.panes[panePackages].View(),
		m.panes[paneBranches].View(),
		m.panes[paneCommits].View(),
	)

	// Right column
	rightCol := m.panes[paneDetail].View()

	// Join columns side by side
	main := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
}

func (m model) renderFooter() string {
	w := m.width

	if m.committing {
		prefix := lipgloss.NewStyle().Foreground(colorGit).Render(" commit: ")
		input := m.commitInput.View()
		line := prefix + input
		return padOrTruncate(line, w)
	}

	if m.statusMsg != "" {
		msg := " " + lipgloss.NewStyle().Foreground(colorHighlight).Render(m.statusMsg)
		return padOrTruncate(msg, w)
	}

	hints := " tab:switch  ↑↓:navigate  1-5:pane  c:commit  p:push  P:pull  q:quit"
	return lipgloss.NewStyle().Foreground(colorDim).Render(padOrTruncate(hints, w))
}

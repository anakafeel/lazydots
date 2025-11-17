package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/anakafeel/LazyDots/internal/config"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

//
// Link status for dotfiles
//

type LinkStatus int

const (
	StatusMissing  LinkStatus = iota // no file at target path
	StatusLinked                     // symlink exists and points to this package file
	StatusConflict                   // file exists but is not the right symlink / some other issue
)

//
// Package list (top-level: FEDORA-WORKSTATION, hypr, kitty, etc.)
//

type packageItem struct {
	name     string
	fullPath string
}

func (p packageItem) Title() string       { return "📦 " + p.name }
func (p packageItem) Description() string { return p.fullPath }
func (p packageItem) FilterValue() string { return p.name }

type packageListModel struct {
	list     list.Model
	rootPath string
	width    int
	height   int
}

func NewPackageListModel(rootPath string, width, height int) packageListModel {
	items := []list.Item{}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		items = append(items, packageItem{
			name:     "❌ Failed to read dotfiles root",
			fullPath: fmt.Sprintf("%v", err),
		})
	} else {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if e.Name() == ".git" {
				// Hide git metadata from the package list
				continue
			}
			full := filepath.Join(rootPath, e.Name())
			items = append(items, packageItem{name: e.Name(), fullPath: full})
		}
		if len(items) == 0 {
			items = append(items, packageItem{
				name:     "⚠️ No packages found",
				fullPath: "Create Stow-style packages in your dotfiles repo (e.g. FEDORA-WORKSTATION/).",
			})
		}
	}

	// Fallback sizes if we haven't received a WindowSizeMsg yet.
	if width == 0 {
		width = 50
	}
	if height == 0 {
		height = 15
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height-2)
	l.Title = fmt.Sprintf("Dotfile Packages in %s", filepath.Base(rootPath))

	return packageListModel{
		list:     l,
		rootPath: rootPath,
		width:    width,
		height:   height,
	}
}

func (m packageListModel) Init() tea.Cmd { return nil }

func (m packageListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Go back to main screen, preserving the configured path.
			return New(config.Config{DotfilesPath: m.rootPath}), nil

		case "enter":
			if it := m.list.SelectedItem(); it != nil {
				if pkg, ok := it.(packageItem); ok {
					// Jump into the file list for this package.
					return NewFileListModel(pkg.fullPath, m.width, m.height), nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m packageListModel) View() string {
	return m.list.View()
}

//
// File list (inside a specific package, recursive)
//

type fileItem struct {
	name   string     // relative path within the package
	status LinkStatus // link status at target path
	target string     // resolved target path under $HOME
}

func (f fileItem) Title() string {
	icon := "❔"
	switch f.status {
	case StatusMissing:
		icon = "⭕" // not linked
	case StatusLinked:
		icon = "✅"
	case StatusConflict:
		icon = "⚠️"
	}
	return icon + " " + f.name
}

func (f fileItem) Description() string {
	if f.target == "" {
		return ""
	}
	return f.target
}

func (f fileItem) FilterValue() string { return f.name }

type fileListModel struct {
	list        list.Model
	packagePath string
	width       int
	height      int
}

func NewFileListModel(packagePath string, width, height int) fileListModel {
	items := []list.Item{}

	home, _ := os.UserHomeDir()

	err := filepath.WalkDir(packagePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory entirely.
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Only list files, not directories.
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(packagePath, path)
		if err != nil {
			rel = path
		}

		// --- Resolve target path under $HOME ---
		// Basic heuristic for Stow-style layout:
		//   e.g. "fish/.config/fish/config.fish" -> "~/.config/fish/config.fish"
		targetRel := rel
		if !strings.HasPrefix(targetRel, ".") {
			if i := strings.IndexRune(targetRel, os.PathSeparator); i != -1 {
				targetRel = targetRel[i+1:]
			}
		}
		targetPath := filepath.Join(home, targetRel)

		status := computeLinkStatus(path, targetPath)

		items = append(items, fileItem{
			name:   rel,
			status: status,
			target: targetPath,
		})

		return nil
	})

	if err != nil {
		items = []list.Item{fileItem{
			name:   fmt.Sprintf("❌ Failed to scan package: %v", err),
			status: StatusConflict,
		}}
	}

	if len(items) == 0 {
		items = append(items, fileItem{
			name:   "⚠️ No files found in this package",
			status: StatusMissing,
		})
	}

	if width == 0 {
		width = 60
	}
	if height == 0 {
		height = 20
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height-2)
	l.Title = fmt.Sprintf("Files in %s", filepath.Base(packagePath))

	return fileListModel{
		list:        l,
		packagePath: packagePath,
		width:       width,
		height:      height,
	}
}

func (m fileListModel) Init() tea.Cmd { return nil }

func (m fileListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Go back to package list.
			root := filepath.Dir(m.packagePath)
			return NewPackageListModel(root, m.width, m.height), nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m fileListModel) View() string {
	return m.list.View()
}

//
// Helpers
//

// computeLinkStatus checks what's at targetPath and whether it is a symlink
// pointing back to srcPath (the file inside the package).
func computeLinkStatus(srcPath, targetPath string) LinkStatus {
	info, err := os.Lstat(targetPath)
	if os.IsNotExist(err) {
		return StatusMissing
	}
	if err != nil {
		return StatusConflict
	}

	// If it's a symlink, check where it points.
	if info.Mode()&os.ModeSymlink != 0 {
		linkDest, err := os.Readlink(targetPath)
		if err != nil {
			return StatusConflict
		}

		absSrc, _ := filepath.Abs(srcPath)
		absDest := linkDest

		// If the symlink is relative, resolve it relative to the directory it's in.
		if !filepath.IsAbs(absDest) {
			dir := filepath.Dir(targetPath)
			absDest = filepath.Join(dir, absDest)
		}
		absDest, _ = filepath.Abs(absDest)

		if absSrc == absDest {
			return StatusLinked
		}
		return StatusConflict
	}

	// Not a symlink – some other file/dir is in the way.
	return StatusConflict
}

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
// 🔹 Link status for dotfiles
//

type LinkStatus int

const (
	StatusMissing  LinkStatus = iota // no file at target path
	StatusLinked                     // symlink exists and points to this package file
	StatusConflict                   // file exists but is not the right symlink / some other issue
)

//
// 🔹 Package list (top-level: FEDORA-WORKSTATION, hypr, kitty, etc.)
//

type packageItem struct {
	name     string
	fullPath string
}

func (p packageItem) Title() string       { return "📦 " + p.name }
func (p packageItem) Description() string { return p.fullPath }
func (p packageItem) FilterValue() string { return p.name }

type packageListModel struct {
	list        list.Model
	rootPath    string
	bannerColor string
	width       int
	height      int
}

func NewPackageListModel(rootPath string, bannerColor string, width, height int) packageListModel {
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
		list:        l,
		rootPath:    rootPath,
		bannerColor: bannerColor,
		width:       width,
		height:      height,
	}
}

func (m packageListModel) Init() tea.Cmd { return nil }

func (m packageListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Go back to main screen, preserving the configured path.
			return New(config.Config{DotfilesPath: m.rootPath}, m.bannerColor), nil

		case "enter":
			if it := m.list.SelectedItem(); it != nil {
				if pkg, ok := it.(packageItem); ok {
					// Jump into the file list for this package.
					return NewFileListModel(pkg.fullPath, m.bannerColor, m.width, m.height), nil
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
// 🔹 File list (inside a specific package, recursive)
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
	bannerColor string
	width       int
	height      int
}

func NewFileListModel(packagePath string, bannerColor string, width, height int) fileListModel {
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
	l.Title = fmt.Sprintf("Files in %s (space: toggle, a/A: link/unlink all)", filepath.Base(packagePath))

	return fileListModel{
		list:        l,
		packagePath: packagePath,
		bannerColor: bannerColor,
		width:       width,
		height:      height,
	}
}

func (m fileListModel) Init() tea.Cmd { return nil }

func (m fileListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Go back to package list.
			root := filepath.Dir(m.packagePath)
			return NewPackageListModel(root, m.bannerColor, m.width, m.height), nil

		case " ", "space":
			// Toggle link/unlink for the selected file (like lazygit's space to stage)
			idx := m.list.Index()
			if idx < 0 || idx >= len(m.list.Items()) {
				break
			}

			it, ok := m.list.Items()[idx].(fileItem)
			if !ok {
				break
			}

			src := filepath.Join(m.packagePath, it.name)
			var err error

			switch it.status {
			case StatusLinked:
				// Currently linked → try to unlink
				err = unlinkDotfile(src, it.target)
			case StatusMissing, StatusConflict:
				// Not linked or conflicting → try to link
				err = linkDotfile(src, it.target)
			}

			// Recompute status after operation
			newStatus := computeLinkStatus(src, it.target)
			it.status = newStatus
			m.list.SetItem(idx, it)

			// Show a status message in the footer (like lazygit)
			if err != nil {
				m.list.NewStatusMessage("⚠️ " + err.Error())
			} else {
				switch newStatus {
				case StatusLinked:
					m.list.NewStatusMessage("✅ Linked " + it.name)
				case StatusMissing:
					m.list.NewStatusMessage("⭕ Unlinked " + it.name)
				case StatusConflict:
					m.list.NewStatusMessage("⚠️ Conflict on " + it.name)
				}
			}

		case "a":
			// Link ALL files in package
			linked, skipped, errors := m.linkAll()
			m.refreshAllStatuses()
			if errors > 0 {
				m.list.NewStatusMessage(fmt.Sprintf("✅ Linked %d, skipped %d, ⚠️ %d errors", linked, skipped, errors))
			} else {
				m.list.NewStatusMessage(fmt.Sprintf("✅ Linked %d files, skipped %d", linked, skipped))
			}

		case "A":
			// Unlink ALL files in package
			unlinked, skipped, errors := m.unlinkAll()
			m.refreshAllStatuses()
			if errors > 0 {
				m.list.NewStatusMessage(fmt.Sprintf("⭕ Unlinked %d, skipped %d, ⚠️ %d errors", unlinked, skipped, errors))
			} else {
				m.list.NewStatusMessage(fmt.Sprintf("⭕ Unlinked %d files, skipped %d", unlinked, skipped))
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m fileListModel) View() string {
	return m.list.View()
}

// linkAll links all files that aren't already linked.
// Returns (linked, skipped, errors) counts.
func (m *fileListModel) linkAll() (int, int, int) {
	linked, skipped, errors := 0, 0, 0
	for i, item := range m.list.Items() {
		it, ok := item.(fileItem)
		if !ok {
			continue
		}
		if it.status == StatusLinked {
			skipped++
			continue
		}
		src := filepath.Join(m.packagePath, it.name)
		if err := linkDotfile(src, it.target); err != nil {
			errors++
		} else {
			linked++
		}
		// Update item status
		it.status = computeLinkStatus(src, it.target)
		m.list.SetItem(i, it)
	}
	return linked, skipped, errors
}

// unlinkAll unlinks all files that are currently linked.
// Returns (unlinked, skipped, errors) counts.
func (m *fileListModel) unlinkAll() (int, int, int) {
	unlinked, skipped, errors := 0, 0, 0
	for i, item := range m.list.Items() {
		it, ok := item.(fileItem)
		if !ok {
			continue
		}
		if it.status != StatusLinked {
			skipped++
			continue
		}
		src := filepath.Join(m.packagePath, it.name)
		if err := unlinkDotfile(src, it.target); err != nil {
			errors++
		} else {
			unlinked++
		}
		// Update item status
		it.status = computeLinkStatus(src, it.target)
		m.list.SetItem(i, it)
	}
	return unlinked, skipped, errors
}

// refreshAllStatuses recomputes status for all items.
func (m *fileListModel) refreshAllStatuses() {
	for i, item := range m.list.Items() {
		it, ok := item.(fileItem)
		if !ok {
			continue
		}
		src := filepath.Join(m.packagePath, it.name)
		it.status = computeLinkStatus(src, it.target)
		m.list.SetItem(i, it)
	}
}

//
// 🔹 Helpers
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

// linkDotfile creates a symlink from srcPath (in the repo) to targetPath (in $HOME).
// It will:
//   - create parent directories if needed
//   - skip if the correct symlink already exists
//   - return an error if a conflicting file/symlink exists
func linkDotfile(srcPath, targetPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	info, err := os.Lstat(targetPath)
	if err == nil {
		// Something already exists at target
		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink; check if it's already correct
			linkDest, err := os.Readlink(targetPath)
			if err != nil {
				return fmt.Errorf("readlink failed: %w", err)
			}

			absSrc, _ := filepath.Abs(srcPath)
			absDest := linkDest
			if !filepath.IsAbs(absDest) {
				absDest = filepath.Join(filepath.Dir(targetPath), absDest)
			}
			absDest, _ = filepath.Abs(absDest)

			if absSrc == absDest {
				// Already correctly linked
				return nil
			}
			return fmt.Errorf("target already linked to a different path: %s", targetPath)
		}

		// Regular file or directory exists
		return fmt.Errorf("target already exists and is not a symlink: %s", targetPath)
	} else if !os.IsNotExist(err) {
		// Some other filesystem error
		return fmt.Errorf("lstat failed: %w", err)
	}

	// Safe to create the symlink
	if err := os.Symlink(srcPath, targetPath); err != nil {
		return fmt.Errorf("symlink failed: %w", err)
	}

	return nil
}

// unlinkDotfile removes a symlink at targetPath IF and only if it points to srcPath.
// It won't touch regular files or symlinks pointing elsewhere.
func unlinkDotfile(srcPath, targetPath string) error {
	info, err := os.Lstat(targetPath)
	if os.IsNotExist(err) {
		// Nothing to do
		return nil
	}
	if err != nil {
		return fmt.Errorf("lstat failed: %w", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("target is not a symlink: %s", targetPath)
	}

	linkDest, err := os.Readlink(targetPath)
	if err != nil {
		return fmt.Errorf("readlink failed: %w", err)
	}

	absSrc, _ := filepath.Abs(srcPath)
	absDest := linkDest
	if !filepath.IsAbs(absDest) {
		absDest = filepath.Join(filepath.Dir(targetPath), absDest)
	}
	absDest, _ = filepath.Abs(absDest)

	if absSrc != absDest {
		return fmt.Errorf("symlink points elsewhere, refusing to remove: %s", targetPath)
	}

	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("remove failed: %w", err)
	}

	return nil
}

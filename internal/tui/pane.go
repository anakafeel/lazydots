package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Pane is the interface every dashboard pane implements.
type Pane interface {
	Update(msg tea.Msg) tea.Cmd
	View() string
	SetSize(width, height int)
	Focus()
	Blur()
	Focused() bool
	Title() string
}

// renderPane wraps content in a bordered box with title in the top border.
// w and h are OUTER dimensions (border included).
func renderPane(title string, content string, w, h int, focused bool) string {
	if w < 4 || h < 3 {
		return ""
	}

	bc, tc := colorBorder, colorTitle
	if focused {
		bc, tc = colorBorderFocus, colorTitleFocus
	}

	bStyle := lipgloss.NewStyle().Foreground(bc)
	tStyle := lipgloss.NewStyle().Foreground(tc).Bold(focused)

	innerW := w - 2
	innerH := h - 2

	// ╭─ Title ──────────╮
	titleText := " " + title + " "
	titleW := lipgloss.Width(titleText)
	dashW := innerW - titleW
	if dashW < 0 {
		dashW = 0
	}
	topLine := bStyle.Render("╭") + tStyle.Render(titleText) + bStyle.Render(strings.Repeat("─", dashW)+"╮")

	// ╰──────────────────╯
	bottomLine := bStyle.Render("╰" + strings.Repeat("─", innerW) + "╯")

	lines := strings.Split(content, "\n")

	var b strings.Builder
	b.WriteString(topLine)
	b.WriteByte('\n')

	for i := 0; i < innerH; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}
		line = padOrTruncate(line, innerW)
		b.WriteString(bStyle.Render("│"))
		b.WriteString(line)
		b.WriteString(bStyle.Render("│"))
		if i < innerH-1 {
			b.WriteByte('\n')
		}
	}

	b.WriteByte('\n')
	b.WriteString(bottomLine)

	return b.String()
}

// padOrTruncate ensures s renders at exactly w visual columns.
func padOrTruncate(s string, w int) string {
	sw := lipgloss.Width(s)
	if sw > w {
		s = lipgloss.NewStyle().MaxWidth(w).Render(s)
		sw = lipgloss.Width(s)
	}
	if sw < w {
		s += strings.Repeat(" ", w-sw)
	}
	return s
}

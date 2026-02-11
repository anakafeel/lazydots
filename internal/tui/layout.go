package tui

// PaneRect holds computed outer dimensions for a single pane.
type PaneRect struct {
	Width  int
	Height int
}

// Layout holds computed dimensions for the entire dashboard.
type Layout struct {
	Status   PaneRect
	Packages PaneRect
	Branches PaneRect
	Commits  PaneRect
	Detail   PaneRect
	FooterH  int
}

// ComputeLayout computes pane dimensions from the terminal size.
//
// Layout (lazygit-inspired):
//
//	┌──────────┬─────────────────┐
//	│ Status   │                 │
//	├──────────┤  Detail/Preview │
//	│ Packages │                 │
//	├──────────┤                 │
//	│ Branches │                 │
//	├──────────┤                 │
//	│ Commits  │                 │
//	└──────────┴─────────────────┘
//	 footer (key hints)
func ComputeLayout(totalW, totalH int) Layout {
	const footerH = 1

	contentH := totalH - footerH
	if contentH < 12 {
		contentH = 12
	}

	// Left column: ~30% of width, clamped
	leftW := totalW * 30 / 100
	if leftW < 30 {
		leftW = 30
	}
	if leftW > totalW/2 {
		leftW = totalW / 2
	}
	rightW := totalW - leftW

	// Status pane: small fixed height
	statusH := 5
	if statusH > contentH/4 {
		statusH = contentH / 4
	}
	if statusH < 3 {
		statusH = 3
	}

	remaining := contentH - statusH
	paneH := remaining / 3
	lastH := remaining - paneH*2 // absorb rounding

	return Layout{
		Status:   PaneRect{Width: leftW, Height: statusH},
		Packages: PaneRect{Width: leftW, Height: paneH},
		Branches: PaneRect{Width: leftW, Height: paneH},
		Commits:  PaneRect{Width: leftW, Height: lastH},
		Detail:   PaneRect{Width: rightW, Height: contentH},
		FooterH:  footerH,
	}
}

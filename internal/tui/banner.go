package tui

import (
	"math/rand/v2"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ASCII logos
const LogoLarge = `
██╗      █████╗ ███████╗██╗   ██╗██████╗  ██████╗ ████████╗███████╗
██║     ██╔══██╗╚══███╔╝╚██╗ ██╔╝██╔══██╗██╔═══██╗╚══██╔══╝██╔════╝
██║     ███████║  ███╔╝  ╚████╔╝ ██║  ██║██║   ██║   ██║   ███████╗
██║     ██╔══██║ ███╔╝    ╚██╔╝  ██║  ██║██║   ██║   ██║   ╚════██║
███████╗██║  ██║███████╗   ██║   ██████╔╝╚██████╔╝   ██║   ███████║
╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═════╝  ╚═════╝    ╚═╝   ╚══════╝`

const LogoCompact = `
 _                  ___       _
| |   __ _ _____  _|   \ ___ | |_ ___
| |__/ _' |_ / || | |) / _ \|  _(_-<
|____\__,_/__|\_, |___/\___/ \__/__/
              |__/`

// Nice terminal-safe ANSI colors
var bannerColors = []string{
	"33",  // blue
	"39",  // cyan
	"63",  // purple
	"99",  // violet
	"111", // light blue
	"141", // lavender
	"171", // magenta
	"205", // pink
	"212", // hot pink
}

// PickBannerColor returns a random color from the palette.
// Call once at startup for consistent branding.
func PickBannerColor() string {
	return bannerColors[rand.IntN(len(bannerColors))]
}

// RenderBanner returns the styled ASCII banner for the given terminal width.
func RenderBanner(width int, color string) string {
	logo := LogoLarge
	if width > 0 && width < 80 {
		logo = LogoCompact
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true)

	return style.Render(strings.TrimLeft(logo, "\n"))
}

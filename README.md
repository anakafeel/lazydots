# LazyDots

A terminal-based dotfile manager inspired by [lazygit](https://github.com/jesseduffield/lazygit). Manage your dotfiles with a simple TUI — visualize symlink status, link/unlink files, and keep your configs organized across machines.

## The Problem

Managing dotfiles is tedious:
- Manually creating symlinks is error-prone
- Tracking what's linked vs what's not is a mental burden
- Git workflows for dotfiles require context-switching to the terminal

LazyDots gives you a visual, interactive way to manage [Stow-style](https://www.gnu.org/software/stow/) dotfile repositories.

## Features

### Implemented
- **Setup wizard** — First-run experience to configure your dotfiles path
- **Package browser** — View Stow-style packages (directories) in your dotfiles repo
- **File browser** — Recursively list files within each package
- **Symlink status** — Visual indicators for each file:
  - ✅ Linked (symlink exists and points to the correct file)
  - ⭕ Missing (not linked)
  - ⚠️ Conflict (file exists but isn't a symlink, or points elsewhere)
- **Toggle linking** — Press `space` to link/unlink individual files
- **Safe operations** — Won't overwrite existing files; only removes symlinks that point to your repo
- **Splash screen** — ASCII logo on startup (skippable with any key or `--no-splash`)

### Planned (Roadmap)
- Batch link/unlink all files in a package
- Git status integration (show uncommitted changes)
- Git commit/push/pull from TUI
- Profile support (different configs for different machines)
- Conflict resolution UI (diff, backup, overwrite options)

## Installation

```bash
go install github.com/anakafeel/LazyDots/cmd/lazydots@latest
```

Or build from source:
```bash
git clone https://github.com/anakafeel/LazyDots.git
cd LazyDots
go build -o lazydots ./cmd/lazydots
```

## Usage

```bash
lazydots              # Launch with splash screen
lazydots --no-splash  # Skip splash screen (useful for scripts)
```

### First Run

On first launch, LazyDots will prompt you to enter your dotfiles path:

```
Enter the path to your dotfiles repository:
> ~/dotfiles
```

The path should point to a Stow-style repository, where each subdirectory is a "package" containing files that mirror your home directory structure:

```
~/dotfiles/
├── fish/
│   └── .config/
│       └── fish/
│           └── config.fish    → ~/.config/fish/config.fish
├── nvim/
│   └── .config/
│       └── nvim/
│           └── init.lua       → ~/.config/nvim/init.lua
└── git/
    └── .gitconfig             → ~/.gitconfig
```

### Keybindings

**Main Screen**
| Key | Action |
|-----|--------|
| `l` | List dotfile packages |
| `r` | Reconfigure dotfiles path |
| `q` | Quit |

**Package List**
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `enter` | Enter package (view files) |
| `/` | Filter packages |
| `q` or `esc` | Back to main screen |

**File List**
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `space` | Toggle link/unlink for selected file |
| `/` | Filter files |
| `q` or `esc` | Back to package list |

## Configuration

Config is stored at:
```
~/.config/lazydots/config.json
```

Contents:
```json
{
  "dotfiles_path": "/home/user/dotfiles"
}
```

You can edit this manually or use `r` in the TUI to reconfigure.

## Stow-Style Layout

LazyDots expects your dotfiles to follow the [GNU Stow](https://www.gnu.org/software/stow/) convention:

- Each top-level directory is a "package" (e.g., `fish/`, `nvim/`, `git/`)
- Files inside packages mirror the structure they should have relative to `$HOME`
- Example: `fish/.config/fish/config.fish` becomes `~/.config/fish/config.fish`

This is the same layout used by many dotfile managers and makes it easy to selectively link groups of configs.

## Project Status

**Work in progress.** Core symlink management is functional. Git integration and profiles are planned.

## Inspiration

- [lazygit](https://github.com/jesseduffield/lazygit) — TUI design and workflow inspiration
- [GNU Stow](https://www.gnu.org/software/stow/) — Symlink farm manager
- [chezmoi](https://www.chezmoi.io/) — Dotfile manager (more feature-rich, but uses templates)

## License

MIT

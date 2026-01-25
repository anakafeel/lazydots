package git

import (
	"os/exec"
	"strconv"
	"strings"
)

// RepoStatus holds git repository status information.
type RepoStatus struct {
	IsRepo      bool
	Branch      string
	Ahead       int
	Behind      int
	Uncommitted int
	HasUpstream bool
}

// GetStatus returns the git status for the given directory.
func GetStatus(repoPath string) RepoStatus {
	status := RepoStatus{}

	// Check if it's a git repo
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return status // Not a git repo
	}
	status.IsRepo = true

	// Get current branch
	cmd = exec.Command("git", "-C", repoPath, "branch", "--show-current")
	if out, err := cmd.Output(); err == nil {
		status.Branch = strings.TrimSpace(string(out))
	}
	if status.Branch == "" {
		status.Branch = "HEAD" // Detached HEAD
	}

	// Get ahead/behind counts (requires upstream)
	cmd = exec.Command("git", "-C", repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if out, err := cmd.Output(); err == nil {
		status.HasUpstream = true
		parts := strings.Fields(strings.TrimSpace(string(out)))
		if len(parts) >= 2 {
			status.Ahead, _ = strconv.Atoi(parts[0])
			status.Behind, _ = strconv.Atoi(parts[1])
		}
	}

	// Get uncommitted changes count
	cmd = exec.Command("git", "-C", repoPath, "status", "--porcelain")
	if out, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			status.Uncommitted = 0
		} else {
			status.Uncommitted = len(lines)
		}
	}

	return status
}

// FormatStatus returns a formatted string for display.
// Example: [main ↑2 ↓0 ●3]
func (s RepoStatus) FormatStatus() string {
	if !s.IsRepo {
		return "not a git repo"
	}

	var parts []string
	parts = append(parts, s.Branch)

	if s.HasUpstream {
		parts = append(parts, "↑"+strconv.Itoa(s.Ahead))
		parts = append(parts, "↓"+strconv.Itoa(s.Behind))
	}

	if s.Uncommitted > 0 {
		parts = append(parts, "●"+strconv.Itoa(s.Uncommitted))
	}

	return "[" + strings.Join(parts, " ") + "]"
}

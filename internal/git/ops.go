package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Commit stages all changes and commits with the given message.
func Commit(repoPath, message string) error {
	// git add -A
	cmd := exec.Command("git", "-C", repoPath, "add", "-A")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s", strings.TrimSpace(string(out)))
	}

	// git commit -m "message"
	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", message)
	if out, err := cmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if strings.Contains(outStr, "nothing to commit") {
			return fmt.Errorf("nothing to commit")
		}
		return fmt.Errorf("git commit failed: %s", outStr)
	}

	return nil
}

// Push pushes to the remote.
func Push(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "push")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Pull pulls from the remote.
func Pull(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

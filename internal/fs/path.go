package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrEmptyPath      = errors.New("path is empty")
	ErrHomeExpand     = errors.New("failed to expand home directory")
	ErrNotExist       = errors.New("path does not exist")
	ErrNotDirectory   = errors.New("path is not a directory")
	ErrInvalidPath    = errors.New("invalid path")
)

// ResolvePath takes user input and returns a normalized absolute path.
// It handles:
//   - Trimming whitespace
//   - Expanding ~ to the user's home directory
//   - Cleaning the path (removing redundant separators, . and ..)
//   - Converting to absolute path
func ResolvePath(input string) (string, error) {
	path := strings.TrimSpace(input)
	if path == "" {
		return "", ErrEmptyPath
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Join(ErrHomeExpand, err)
		}

		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(home, path[2:])
		} else {
			// Handle ~username format (not supported, return error)
			return "", errors.New("~username expansion is not supported; use absolute path or ~/")
		}
	}

	// Clean the path (removes redundant separators, . and ..)
	path = filepath.Clean(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errors.Join(ErrInvalidPath, err)
	}

	return absPath, nil
}

// ValidateDirectory checks that the given path exists and is a directory.
// It expects an already-resolved absolute path (from ResolvePath).
func ValidateDirectory(absPath string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotExist
		}
		return err
	}

	if !info.IsDir() {
		return ErrNotDirectory
	}

	return nil
}

// ResolveAndValidateDirectory is a convenience function that resolves
// the input path and validates it's an existing directory.
func ResolveAndValidateDirectory(input string) (string, error) {
	absPath, err := ResolvePath(input)
	if err != nil {
		return "", err
	}

	if err := ValidateDirectory(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

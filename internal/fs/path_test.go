package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: ErrEmptyPath,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    "",
			wantErr: ErrEmptyPath,
		},
		{
			name:  "tilde only",
			input: "~",
			want:  home,
		},
		{
			name:  "tilde with path",
			input: "~/Documents",
			want:  filepath.Join(home, "Documents"),
		},
		{
			name:  "tilde with nested path",
			input: "~/foo/bar/baz",
			want:  filepath.Join(home, "foo/bar/baz"),
		},
		{
			name:  "tilde with whitespace",
			input: "  ~/Documents  ",
			want:  filepath.Join(home, "Documents"),
		},
		{
			name:  "absolute path",
			input: "/usr/local/bin",
			want:  "/usr/local/bin",
		},
		{
			name:  "relative path",
			input: "foo/bar",
			want:  filepath.Join(cwd, "foo/bar"),
		},
		{
			name:  "dot path",
			input: "./foo",
			want:  filepath.Join(cwd, "foo"),
		},
		{
			name:  "path with redundant separators",
			input: "~/foo//bar///baz",
			want:  filepath.Join(home, "foo/bar/baz"),
		},
		{
			name:  "path with dot segments",
			input: "~/foo/./bar/../baz",
			want:  filepath.Join(home, "foo/baz"),
		},
		{
			name:    "tilde username not supported",
			input:   "~otheruser/path",
			want:    "",
			wantErr: nil, // We check for non-nil error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePath(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ResolvePath(%q) expected error %v, got nil", tt.input, tt.wantErr)
					return
				}
				return
			}

			// Special case: ~username should return an error
			if tt.name == "tilde username not supported" {
				if err == nil {
					t.Errorf("ResolvePath(%q) expected error for ~username, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolvePath(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got != tt.want {
				t.Errorf("ResolvePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "lazydots-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temp file
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "valid directory",
			path:    tmpDir,
			wantErr: nil,
		},
		{
			name:    "non-existent path",
			path:    "/this/path/does/not/exist/at/all",
			wantErr: ErrNotExist,
		},
		{
			name:    "path is a file",
			path:    tmpFile,
			wantErr: ErrNotDirectory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectory(tt.path)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateDirectory(%q) unexpected error: %v", tt.path, err)
				}
				return
			}

			if err == nil {
				t.Errorf("ValidateDirectory(%q) expected error %v, got nil", tt.path, tt.wantErr)
				return
			}
		})
	}
}

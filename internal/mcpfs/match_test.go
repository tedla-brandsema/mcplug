package mcpfs

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestMatcherAllowsIncludedFile(t *testing.T) {
	root := t.TempDir()

	m, err := NewMatcher(root, []string{"**/*.go"}, nil, false, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if !m.AllowFile("cmd/mcpfs/main.go") {
		t.Fatal("expected go file to be allowed")
	}
}

func TestMatcherRejectsFileOutsideIncludes(t *testing.T) {
	root := t.TempDir()

	m, err := NewMatcher(root, []string{"**/*.go"}, nil, false, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if m.AllowFile("notes.txt") {
		t.Fatal("expected txt file to be rejected")
	}
}

func TestMatcherRejectsExplicitExclude(t *testing.T) {
	root := t.TempDir()

	m, err := NewMatcher(root, []string{"**/*.go"}, []string{"**/secret.go"}, false, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if m.AllowFile("internal/secret.go") {
		t.Fatal("expected explicit exclude to reject file")
	}
}

func TestMatcherAllowsDirsForTraversalWhenIncludesExist(t *testing.T) {
	root := t.TempDir()

	m, err := NewMatcher(root, []string{"**/*.go"}, nil, false, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if !m.AllowDir("internal/mcpfs") {
		t.Fatal("expected dir traversal to be allowed")
	}
}

func TestMatcherAppliesRootGitignore(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("tmp/\n*.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := NewMatcher(root, nil, nil, true, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if m.AllowFile("tmp/file.txt") {
		t.Fatal("expected tmp/file.txt to be ignored")
	}

	if m.AllowFile("app.log") {
		t.Fatal("expected app.log to be ignored")
	}

	if !m.AllowFile("README.md") {
		t.Fatal("expected README.md to be allowed")
	}
}

func TestMatcherAppliesNestedGitignore(t *testing.T) {
	root := t.TempDir()

	nested := filepath.Join(root, "internal")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(nested, ".gitignore"), []byte("generated/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := NewMatcher(root, nil, nil, true, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if m.AllowFile("internal/generated/file.go") {
		t.Fatal("expected nested gitignore rule to reject file")
	}

	if !m.AllowFile("generated/file.go") {
		t.Fatal("expected root generated/file.go to be allowed")
	}
}

func TestMatcherSupportsGitignoreNegation(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n!important.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := NewMatcher(root, nil, nil, true, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if m.AllowFile("debug.log") {
		t.Fatal("expected debug.log to be ignored")
	}

	if !m.AllowFile("important.log") {
		t.Fatal("expected important.log to be re-included")
	}
}
package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveInsideRootRejectsAbsolutePath(t *testing.T) {
	root := t.TempDir()

	_, err := ResolveInsideRoot(root, filepath.Join(root, "file.txt"))
	if err == nil {
		t.Fatal("expected absolute path to be rejected")
	}
}

func TestResolveInsideRootRejectsDotDotEscape(t *testing.T) {
	root := t.TempDir()

	_, err := ResolveInsideRoot(root, "../outside.txt")
	if err == nil {
		t.Fatal("expected dot-dot escape to be rejected")
	}
}

func TestResolveInsideRootAllowsNestedFile(t *testing.T) {
	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "b", "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveInsideRoot(root, filepath.Join("a", "b", "file.txt"))
	if err != nil {
		t.Fatalf("expected nested file to resolve: %v", err)
	}

	want := filepath.Join(root, "a", "b", "file.txt")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveInsideRootRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires privileges on some Windows setups")
	}

	root := t.TempDir()
	outside := t.TempDir()

	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "link.txt")); err != nil {
		t.Fatal(err)
	}

	_, err := ResolveInsideRoot(root, "link.txt")
	if err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}

func TestResolveInsideRootAllowsRoot(t *testing.T) {
	root := t.TempDir()

	got, err := ResolveInsideRoot(root, ".")
	if err != nil {
		t.Fatalf("expected root to resolve: %v", err)
	}

	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}

	if got != realRoot {
		t.Fatalf("got %q, want %q", got, realRoot)
	}
}

func TestResolveWritableInsideRootAllowsNewFile(t *testing.T) {
	root := t.TempDir()

	got, err := ResolveWritableInsideRoot(root, filepath.Join("a", "b", "file.txt"))
	if err != nil {
		t.Fatalf("ResolveWritableInsideRoot returned error: %v", err)
	}

	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(realRoot, "a", "b", "file.txt")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveWritableInsideRootRejectsRoot(t *testing.T) {
	root := t.TempDir()

	_, err := ResolveWritableInsideRoot(root, ".")
	if err == nil {
		t.Fatal("expected root path to be rejected")
	}
}

func TestResolveWritableInsideRootRejectsSymlinkParentEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires privileges on some Windows setups")
	}

	root := t.TempDir()
	outside := t.TempDir()

	if err := os.Symlink(outside, filepath.Join(root, "link-outside")); err != nil {
		t.Fatal(err)
	}

	_, err := ResolveWritableInsideRoot(root, filepath.Join("link-outside", "new.txt"))
	if err == nil {
		t.Fatal("expected symlink parent escape to be rejected")
	}
}

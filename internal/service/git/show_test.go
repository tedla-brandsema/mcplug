package git

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
)

func TestParseShowCommit(t *testing.T) {
	input := "abc123def\x00abc123d\x00Tedla\x00tedla@example.com\x002026-05-05T08:00:00+02:00\x00Add git show\x00Body text"

	commit, err := ParseShowCommit(input)
	if err != nil {
		t.Fatal(err)
	}

	if commit.Hash != "abc123def" {
		t.Fatalf("hash = %q", commit.Hash)
	}
	if commit.ShortHash != "abc123d" {
		t.Fatalf("short_hash = %q", commit.ShortHash)
	}
	if commit.AuthorName != "Tedla" {
		t.Fatalf("author_name = %q", commit.AuthorName)
	}
	if commit.AuthorEmail != "tedla@example.com" {
		t.Fatalf("author_email = %q", commit.AuthorEmail)
	}
	if commit.AuthorDate != "2026-05-05T08:00:00+02:00" {
		t.Fatalf("author_date = %q", commit.AuthorDate)
	}
	if commit.Subject != "Add git show" {
		t.Fatalf("subject = %q", commit.Subject)
	}
	if commit.Body != "Body text" {
		t.Fatalf("body = %q", commit.Body)
	}
}

func TestParseShowCommitTrimsBody(t *testing.T) {
	input := "hash\x00short\x00Alice\x00alice@example.com\x002026-05-05T08:00:00+02:00\x00Subject\x00\n\nBody text\n\n"

	commit, err := ParseShowCommit(input)
	if err != nil {
		t.Fatal(err)
	}

	if commit.Body != "Body text" {
		t.Fatalf("body = %q", commit.Body)
	}
}

func TestParseShowCommitInvalid(t *testing.T) {
	_, err := ParseShowCommit("too few\x00fields")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateShowRev(t *testing.T) {
	tests := []struct {
		name    string
		rev     string
		wantErr bool
	}{
		{name: "head", rev: "HEAD"},
		{name: "head parent", rev: "HEAD~1"},
		{name: "hash", rev: "abc123def"},
		{name: "tag", rev: "v0.2.0"},
		{name: "empty", rev: "", wantErr: true},
		{name: "spaces", rev: "   ", wantErr: true},
		{name: "leading whitespace", rev: " HEAD", wantErr: true},
		{name: "trailing whitespace", rev: "HEAD ", wantErr: true},
		{name: "option like", rev: "--help", wantErr: true},
		{name: "nul", rev: "HEAD\x00bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateShowRev(tt.rev)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestShowHonorsMaxBytes(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	writeFile(t, dir, "large.txt", strings.Repeat("0123456789\n", 200))
	gitCommand(t, dir, "add", "large.txt")
	gitCommand(t, dir, "commit", "-m", "add large file")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Show(context.Background(), ShowArgs{
		RootID:   "repo",
		Rev:      "HEAD",
		MaxBytes: 80,
	})
	if err != nil {
		t.Fatalf("Show returned error: %v", err)
	}

	if result.Bytes > 80 {
		t.Fatalf("Bytes = %d, want <= 80", result.Bytes)
	}
	if len(result.Diff) > 80 {
		t.Fatalf("len(Diff) = %d, want <= 80", len(result.Diff))
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestCapStringBytes(t *testing.T) {
	got, truncated := capStringBytes("hello world", 5)
	if got != "hello" {
		t.Fatalf("got = %q, want %q", got, "hello")
	}
	if !truncated {
		t.Fatal("truncated = false, want true")
	}
}

func TestCapStringBytesNoTruncation(t *testing.T) {
	got, truncated := capStringBytes("hello", 10)
	if got != "hello" {
		t.Fatalf("got = %q, want %q", got, "hello")
	}
	if truncated {
		t.Fatal("truncated = true, want false")
	}
}

func initTestRepo(t *testing.T, dir string) {
	t.Helper()

	gitCommand(t, dir, "init")
	gitCommand(t, dir, "config", "user.name", "Test User")
	gitCommand(t, dir, "config", "user.email", "test@example.com")
}

func gitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func newTestService(t *testing.T, configs ...config.RootConfig) *Service {
	t.Helper()

	roots := make([]*core.Root, 0, len(configs))
	for _, cfg := range configs {
		root, err := core.NewRoot(cfg, discardLogger())
		if err != nil {
			t.Fatalf("NewRoot(%q) returned error: %v", cfg.ID, err)
		}
		roots = append(roots, root)
	}

	return New(roots, discardLogger())
}

func testRootConfig(id string, dir string) config.RootConfig {
	return config.RootConfig{
		ID:           id,
		Path:         dir,
		Mode:         config.ModeRead,
		Include:      []string{"**/*"},
		Exclude:      nil,
		UseGitignore: false,
		MaxFileBytes: 262144,
	}
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

package fs

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
)

func TestRootsReturnsRootsSortedByID(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	svc := newTestService(t,
		testRootConfig("z-root", dirB),
		testRootConfig("a-root", dirA),
	)

	result, err := svc.Roots(context.Background(), RootsArgs{})
	if err != nil {
		t.Fatalf("Roots returned error: %v", err)
	}

	if len(result.Roots) != 2 {
		t.Fatalf("len(Roots) = %d, want 2", len(result.Roots))
	}
	if result.Roots[0].ID != "a-root" {
		t.Fatalf("Roots[0].ID = %q, want %q", result.Roots[0].ID, "a-root")
	}
	if result.Roots[1].ID != "z-root" {
		t.Fatalf("Roots[1].ID = %q, want %q", result.Roots[1].ID, "z-root")
	}
}

func TestListHidesExcludedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "visible.txt", "visible")
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	result, err := svc.List(context.Background(), ListArgs{
		RootID: "repo",
		Path:   ".",
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	assertEntryPaths(t, result.Entries, []string{"visible.txt"})
}

func TestListRecursiveHonorsMaxEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "b.txt", "b")
	writeFile(t, dir, "nested/c.txt", "c")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.List(context.Background(), ListArgs{
		RootID:     "repo",
		Path:       ".",
		Recursive:  true,
		MaxEntries: 2,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(result.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(result.Entries))
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestReadReturnsContentAndMetadata(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "hello.txt" {
		t.Fatalf("Path = %q, want %q", result.Path, "hello.txt")
	}
	if result.Content != "hello world" {
		t.Fatalf("Content = %q, want %q", result.Content, "hello world")
	}
	if result.Bytes != len("hello world") {
		t.Fatalf("Bytes = %d, want %d", result.Bytes, len("hello world"))
	}
	if result.Size != int64(len("hello world")) {
		t.Fatalf("Size = %d, want %d", result.Size, len("hello world"))
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestReadHonorsOffsetAndLimit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "abcdef")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "data.txt",
		Offset: 2,
		Limit:  3,
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if result.Content != "cde" {
		t.Fatalf("Content = %q, want %q", result.Content, "cde")
	}
	if result.Offset != 2 {
		t.Fatalf("Offset = %d, want 2", result.Offset)
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestReadOffsetPastEOFReturnsEmptyContent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "abc")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "data.txt",
		Offset: 99,
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if result.Content != "" {
		t.Fatalf("Content = %q, want empty", result.Content)
	}
	if result.Bytes != 0 {
		t.Fatalf("Bytes = %d, want 0", result.Bytes)
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestReadRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "secret.txt",
	})
	if err == nil {
		t.Fatal("Read returned nil error")
	}
}

func TestReadRejectsFileOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "large.txt", "abcdef")

	cfg := testRootConfig("repo", dir)
	cfg.MaxFileBytes = 3

	svc := newTestService(t, cfg)

	_, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "large.txt",
	})
	if err == nil {
		t.Fatal("Read returned nil error")
	}
}

func TestReadRejectsNegativeOffset(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "abc")

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Read(context.Background(), ReadArgs{
		RootID: "repo",
		Path:   "data.txt",
		Offset: -1,
	})
	if err == nil {
		t.Fatal("Read returned nil error")
	}
}

func TestSearchFindsMatchingLines(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "alpha\nneedle here\nomega\n")
	writeFile(t, dir, "b.txt", "nothing\nneedle again\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Search(context.Background(), SearchArgs{
		RootID: "repo",
		Query:  "needle",
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(result.Matches) != 2 {
		t.Fatalf("len(Matches) = %d, want 2", len(result.Matches))
	}
	if result.Matches[0].Line != 2 {
		t.Fatalf("Matches[0].Line = %d, want 2", result.Matches[0].Line)
	}
}

func TestSearchHonorsGlob(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "needle\n")
	writeFile(t, dir, "b.md", "needle\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Search(context.Background(), SearchArgs{
		RootID: "repo",
		Query:  "needle",
		Glob:   "**/*.md",
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(result.Matches) != 1 {
		t.Fatalf("len(Matches) = %d, want 1", len(result.Matches))
	}
	if result.Matches[0].Path != "b.md" {
		t.Fatalf("Matches[0].Path = %q, want %q", result.Matches[0].Path, "b.md")
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	dir := t.TempDir()

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Search(context.Background(), SearchArgs{
		RootID: "repo",
		Query:  "",
	})
	if err == nil {
		t.Fatal("Search returned nil error")
	}
}

func TestSearchTruncatesAtMaxResults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "needle 1\nneedle 2\nneedle 3\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Search(context.Background(), SearchArgs{
		RootID:     "repo",
		Query:      "needle",
		MaxResults: 2,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(result.Matches) != 2 {
		t.Fatalf("len(Matches) = %d, want 2", len(result.Matches))
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestStatReturnsFileMetadata(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Stat(context.Background(), StatArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "hello.txt" {
		t.Fatalf("Path = %q, want %q", result.Path, "hello.txt")
	}
	if result.Type != "file" {
		t.Fatalf("Type = %q, want %q", result.Type, "file")
	}
	if result.Size != 5 {
		t.Fatalf("Size = %d, want 5", result.Size)
	}
	if result.MTime == "" {
		t.Fatal("MTime is empty")
	}
	if result.Mode == "" {
		t.Fatal("Mode is empty")
	}
}

func TestStatReturnsDirectoryMetadata(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, dir, "nested")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Stat(context.Background(), StatArgs{
		RootID: "repo",
		Path:   "nested",
	})
	if err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}

	if result.Type != "dir" {
		t.Fatalf("Type = %q, want %q", result.Type, "dir")
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

func mkdir(t *testing.T, root string, rel string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", path, err)
	}
}

func assertEntryPaths(t *testing.T, entries []Entry, want []string) {
	t.Helper()

	if len(entries) != len(want) {
		t.Fatalf("len(entries) = %d, want %d; entries = %#v", len(entries), len(want), entries)
	}

	for i := range want {
		if entries[i].Path != want[i] {
			t.Fatalf("entries[%d].Path = %q, want %q; entries = %#v", i, entries[i].Path, want[i], entries)
		}
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

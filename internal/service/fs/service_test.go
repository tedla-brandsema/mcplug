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

func TestTreeReturnsStructuredTreeAndText(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "nested/b.txt", "b")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:   "repo",
		Path:     ".",
		MaxDepth: 3,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "." {
		t.Fatalf("Path = %q, want %q", result.Path, ".")
	}
	if result.Root.Path != "." {
		t.Fatalf("Root.Path = %q, want %q", result.Root.Path, ".")
	}
	if result.Root.Type != "dir" {
		t.Fatalf("Root.Type = %q, want %q", result.Root.Type, "dir")
	}

	assertTreeEntryPaths(t, result.Entries, []string{"a.txt", "nested", "nested/b.txt"})

	if result.Entries[0].Depth != 1 {
		t.Fatalf("Entries[0].Depth = %d, want 1", result.Entries[0].Depth)
	}
	if result.Entries[2].Depth != 2 {
		t.Fatalf("Entries[2].Depth = %d, want 2", result.Entries[2].Depth)
	}
	if result.Entries[2].ParentPath != "nested" {
		t.Fatalf("Entries[2].ParentPath = %q, want %q", result.Entries[2].ParentPath, "nested")
	}

	wantText := ".\n├── a.txt\n└── nested\n    └── b.txt"
	if result.Text != wantText {
		t.Fatalf("Text = %q, want %q", result.Text, wantText)
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
	}
}

func TestTreeHonorsMaxDepth(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "nested/deeper/file.txt", "x")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:   "repo",
		Path:     ".",
		MaxDepth: 1,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"nested"})
}

func TestTreeHonorsMaxEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "b.txt", "b")
	writeFile(t, dir, "c.txt", "c")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:     "repo",
		Path:       ".",
		MaxEntries: 2,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"a.txt", "b.txt"})
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestTreeCanExcludeFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "nested/b.txt", "b")

	includeFiles := false
	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID:       "repo",
		Path:         ".",
		IncludeFiles: &includeFiles,
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"nested"})
}

func TestTreeHidesExcludedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "visible.txt", "visible")
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	result, err := svc.Tree(context.Background(), TreeArgs{
		RootID: "repo",
		Path:   ".",
	})
	if err != nil {
		t.Fatalf("Tree returned error: %v", err)
	}

	assertTreeEntryPaths(t, result.Entries, []string{"visible.txt"})
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

func TestReadLinesReturnsLineRange(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\ntwo\nthree\nfour\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:    "repo",
		Path:      "data.txt",
		StartLine: 2,
		EndLine:   3,
	})
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want %q", result.RootID, "repo")
	}
	if result.Path != "data.txt" {
		t.Fatalf("Path = %q, want %q", result.Path, "data.txt")
	}
	if result.StartLine != 2 {
		t.Fatalf("StartLine = %d, want 2", result.StartLine)
	}
	if result.EndLine != 3 {
		t.Fatalf("EndLine = %d, want 3", result.EndLine)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("len(Lines) = %d, want 2", len(result.Lines))
	}
	if result.Lines[0].Number != 2 || result.Lines[0].Text != "two" {
		t.Fatalf("Lines[0] = %#v, want line 2 two", result.Lines[0])
	}
	if result.Lines[1].Number != 3 || result.Lines[1].Text != "three" {
		t.Fatalf("Lines[1] = %#v, want line 3 three", result.Lines[1])
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestReadLinesDefaultsStartLine(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\ntwo\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:  "repo",
		Path:    "data.txt",
		EndLine: 1,
	})
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	if len(result.Lines) != 1 {
		t.Fatalf("len(Lines) = %d, want 1", len(result.Lines))
	}
	if result.Lines[0].Number != 1 || result.Lines[0].Text != "one" {
		t.Fatalf("Lines[0] = %#v, want line 1 one", result.Lines[0])
	}
}

func TestReadLinesRejectsInvalidRange(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:    "repo",
		Path:      "data.txt",
		StartLine: 3,
		EndLine:   2,
	})
	if err == nil {
		t.Fatal("ReadLines returned nil error")
	}
}

func TestReadLinesRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.txt", "secret\n")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID: "repo",
		Path:   "secret.txt",
	})
	if err == nil {
		t.Fatal("ReadLines returned nil error")
	}
}

func TestReadLinesRejectsFileOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "large.txt", "abcdef\n")

	cfg := testRootConfig("repo", dir)
	cfg.MaxFileBytes = 3

	svc := newTestService(t, cfg)

	_, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID: "repo",
		Path:   "large.txt",
	})
	if err == nil {
		t.Fatal("ReadLines returned nil error")
	}
}

func TestReadLinesReturnsEmptyPastEOF(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "data.txt", "one\ntwo\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.ReadLines(context.Background(), ReadLinesArgs{
		RootID:    "repo",
		Path:      "data.txt",
		StartLine: 10,
		EndLine:   20,
	})
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	if len(result.Lines) != 0 {
		t.Fatalf("len(Lines) = %d, want 0", len(result.Lines))
	}
	if result.Truncated {
		t.Fatal("Truncated = true, want false")
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

func assertTreeEntryPaths(t *testing.T, entries []TreeEntry, want []string) {
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

package fs

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

func TestWriteRejectsReadRoot(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:  "repo",
		Path:    "hello.txt",
		Content: "hello",
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}
}

func TestWriteCreatesFileInReadWriteRoot(t *testing.T) {
	dir := t.TempDir()
	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Write(context.Background(), WriteArgs{
		RootID:  "repo",
		Path:    "hello.txt",
		Content: "hello world",
	})
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want repo", result.RootID)
	}
	if result.Path != "hello.txt" {
		t.Fatalf("Path = %q, want hello.txt", result.Path)
	}
	if result.Bytes != len("hello world") {
		t.Fatalf("Bytes = %d, want %d", result.Bytes, len("hello world"))
	}
	if result.Mode != string(config.ModeReadWrite) {
		t.Fatalf("Mode = %q, want %q", result.Mode, config.ModeReadWrite)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Fatalf("file content = %q, want hello world", data)
	}
}

func TestWriteReplacesExistingFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "old")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:  "repo",
		Path:    "hello.txt",
		Content: "new",
	})
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("file content = %q, want new", data)
	}
}

func TestWriteWithExpectedSHA256ReplacesMatchingExistingFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "old")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:         "repo",
		Path:           "hello.txt",
		Content:        "new",
		ExpectedSHA256: testSHA256("old"),
	})
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("file content = %q, want new", data)
	}
}

func TestWriteWithExpectedSHA256RejectsMismatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "old")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:         "repo",
		Path:           "hello.txt",
		Content:        "new",
		ExpectedSHA256: testSHA256("different"),
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old" {
		t.Fatalf("file content = %q, want old", data)
	}
}

func TestWriteWithExpectedSHA256RejectsMissingFile(t *testing.T) {
	dir := t.TempDir()

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:         "repo",
		Path:           "hello.txt",
		Content:        "new",
		ExpectedSHA256: testSHA256("old"),
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}
}

func TestWriteCreatesParentDirsWhenRequested(t *testing.T) {
	dir := t.TempDir()

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:     "repo",
		Path:       "a/b/hello.txt",
		Content:    "hello",
		CreateDirs: true,
	})
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "a", "b", "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("file content = %q, want hello", data)
	}
}

func TestWriteRejectsMissingParentWithoutCreateDirs(t *testing.T) {
	dir := t.TempDir()

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:  "repo",
		Path:    "a/b/hello.txt",
		Content: "hello",
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}
}

func TestWriteRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:  "repo",
		Path:    "secret.txt",
		Content: "secret",
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}
}

func TestWriteRejectsContentOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite
	cfg.MaxFileBytes = 3

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:  "repo",
		Path:    "large.txt",
		Content: "abcd",
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}
}

func TestWriteRejectsDirectoryTarget(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, dir, "target")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:  "repo",
		Path:    "target",
		Content: "hello",
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}
}

func TestWriteRejectsSymlinkParentEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires privileges on some Windows setups")
	}

	dir := t.TempDir()
	outside := t.TempDir()

	if err := os.Symlink(outside, filepath.Join(dir, "link-outside")); err != nil {
		t.Fatal(err)
	}

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Write(context.Background(), WriteArgs{
		RootID:     "repo",
		Path:       "link-outside/new.txt",
		Content:    "hello",
		CreateDirs: true,
	})
	if err == nil {
		t.Fatal("Write returned nil error")
	}
}

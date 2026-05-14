package fs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

func TestHashReturnsFileMetadata(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Hash(context.Background(), HashArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want repo", result.RootID)
	}
	if result.Path != "hello.txt" {
		t.Fatalf("Path = %q, want hello.txt", result.Path)
	}
	if result.Size != int64(len("hello world")) {
		t.Fatalf("Size = %d, want %d", result.Size, len("hello world"))
	}
	if result.SHA256 != testSHA256("hello world") {
		t.Fatalf("SHA256 = %q, want %q", result.SHA256, testSHA256("hello world"))
	}
	if result.MTime == "" {
		t.Fatal("MTime is empty")
	}
	if result.Mode == "" {
		t.Fatal("Mode is empty")
	}
}

func TestHashRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.Hash(context.Background(), HashArgs{
		RootID: "repo",
		Path:   "secret.txt",
	})
	if err == nil {
		t.Fatal("Hash returned nil error")
	}
}

func TestHashRejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, dir, "target")

	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Hash(context.Background(), HashArgs{
		RootID: "repo",
		Path:   "target",
	})
	if err == nil {
		t.Fatal("Hash returned nil error")
	}
}

func TestHashRejectsFileOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "large.txt", "abcd")

	cfg := testRootConfig("repo", dir)
	cfg.MaxFileBytes = 3

	svc := newTestService(t, cfg)

	_, err := svc.Hash(context.Background(), HashArgs{
		RootID: "repo",
		Path:   "large.txt",
	})
	if err == nil {
		t.Fatal("Hash returned nil error")
	}
}

func TestHashWorksForReadWriteRoot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Hash(context.Background(), HashArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}
	if result.SHA256 != testSHA256("hello") {
		t.Fatalf("SHA256 = %q, want %q", result.SHA256, testSHA256("hello"))
	}
}

func testSHA256(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

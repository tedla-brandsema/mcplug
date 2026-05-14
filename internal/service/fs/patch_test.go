package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

func TestPatchRejectsReadRoot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world")
	svc := newTestService(t, testRootConfig("repo", dir))

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchReplacesExactText(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	if !result.Changed {
		t.Fatal("Changed = false, want true")
	}
	if result.EditsApplied != 1 {
		t.Fatalf("EditsApplied = %d, want 1", result.EditsApplied)
	}
	if result.BytesBefore != len("hello world\n") {
		t.Fatalf("BytesBefore = %d, want %d", result.BytesBefore, len("hello world\n"))
	}
	if result.BytesAfter != len("goodbye world\n") {
		t.Fatalf("BytesAfter = %d, want %d", result.BytesAfter, len("goodbye world\n"))
	}
	if !strings.Contains(result.Diff, "-hello world\n") {
		t.Fatalf("Diff = %q, want removed line", result.Diff)
	}
	if !strings.Contains(result.Diff, "+goodbye world\n") {
		t.Fatalf("Diff = %q, want added line", result.Diff)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "goodbye world\n" {
		t.Fatalf("file content = %q, want goodbye world", data)
	}
}

func TestPatchDryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		DryRun: true,
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}

	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if !result.Changed {
		t.Fatal("Changed = false, want true")
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world\n" {
		t.Fatalf("file content = %q, want unchanged", data)
	}
}

func TestPatchAppliesMultipleEdits(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "one two three\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "one", New: "ONE"},
			{Old: "three", New: "THREE"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}
	if result.EditsApplied != 2 {
		t.Fatalf("EditsApplied = %d, want 2", result.EditsApplied)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "ONE two THREE\n" {
		t.Fatalf("file content = %q, want patched", data)
	}
}

func TestPatchRejectsZeroMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "missing", New: "replacement"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsMultipleMatches(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello hello\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsEmptyOldText(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "", New: "replacement"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsEmptyEdits(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsExcludedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.txt", "secret")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite
	cfg.Exclude = []string{"secret.txt"}

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "secret.txt",
		Edits: []PatchEdit{
			{Old: "secret", New: "public"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsContentOverMaxFileBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite
	cfg.MaxFileBytes = 5

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "hello", New: "hello world"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchRejectsDirectoryTarget(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, dir, "target")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "target",
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}
}

func TestPatchMultiEditFailureIsAtomic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "one two three\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	_, err := svc.Patch(context.Background(), PatchArgs{
		RootID: "repo",
		Path:   "hello.txt",
		Edits: []PatchEdit{
			{Old: "one", New: "ONE"},
			{Old: "missing", New: "MISSING"},
		},
	})
	if err == nil {
		t.Fatal("Patch returned nil error")
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "one two three\n" {
		t.Fatalf("file content = %q, want unchanged", data)
	}
}

func TestPatchTruncatesDiff(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello world\n")

	cfg := testRootConfig("repo", dir)
	cfg.Mode = config.ModeReadWrite

	svc := newTestService(t, cfg)

	result, err := svc.Patch(context.Background(), PatchArgs{
		RootID:       "repo",
		Path:         "hello.txt",
		DryRun:       true,
		MaxDiffBytes: 10,
		Edits: []PatchEdit{
			{Old: "hello", New: "goodbye"},
		},
	})
	if err != nil {
		t.Fatalf("Patch returned error: %v", err)
	}
	if !result.DiffTruncated {
		t.Fatal("DiffTruncated = false, want true")
	}
	if len(result.Diff) > 10 {
		t.Fatalf("len(Diff) = %d, want <= 10", len(result.Diff))
	}
}

package project

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
	fsservice "github.com/tedla-brandsema/mcpfs/internal/service/fs"
	gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"
)

func TestOverviewReturnsProjectSummary(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	writeFile(t, dir, "README.md", "# Project\n")
	writeFile(t, dir, "TODO.md", "# TODO\n")
	writeFile(t, dir, "go.mod", "module example.com/project\n")
	writeFile(t, dir, "internal/app.go", "package internal\n")
	writeFile(t, dir, "internal/app_test.go", "package internal\n")

	gitCommand(t, dir, "add", ".")
	gitCommand(t, dir, "commit", "-m", "initial commit")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Overview(context.Background(), OverviewArgs{
		RootID:        "repo",
		Path:          ".",
		MaxDepth:      2,
		MaxEntries:    100,
		RecentCommits: 3,
	})
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}

	if result.RootID != "repo" {
		t.Fatalf("RootID = %q, want repo", result.RootID)
	}
	if result.Path != "." {
		t.Fatalf("Path = %q, want .", result.Path)
	}
	if result.MaxDepth != 2 {
		t.Fatalf("MaxDepth = %d, want 2", result.MaxDepth)
	}
	if result.MaxEntries != 100 {
		t.Fatalf("MaxEntries = %d, want 100", result.MaxEntries)
	}
	if result.TreeText == "" {
		t.Fatal("TreeText is empty")
	}
	if len(result.TopLevel) == 0 {
		t.Fatal("TopLevel is empty")
	}
	if !containsString(result.ImportantFiles, "README.md") {
		t.Fatalf("ImportantFiles = %#v, want README.md", result.ImportantFiles)
	}
	if !containsString(result.ImportantFiles, "TODO.md") {
		t.Fatalf("ImportantFiles = %#v, want TODO.md", result.ImportantFiles)
	}
	if !containsString(result.ImportantFiles, "go.mod") {
		t.Fatalf("ImportantFiles = %#v, want go.mod", result.ImportantFiles)
	}
	if result.Counts.Files == 0 {
		t.Fatal("Counts.Files = 0, want > 0")
	}
	if result.Counts.TestFiles == 0 {
		t.Fatal("Counts.TestFiles = 0, want > 0")
	}
	if result.Counts.DocumentationFiles == 0 {
		t.Fatal("Counts.DocumentationFiles = 0, want > 0")
	}
	if result.Counts.ConfigurationFiles == 0 {
		t.Fatal("Counts.ConfigurationFiles = 0, want > 0")
	}
	if !result.Git.Available {
		t.Fatalf("Git.Available = false, error = %q", result.Git.Error)
	}
	if !result.Git.Clean {
		t.Fatal("Git.Clean = false, want true")
	}
	if len(result.Git.RecentCommits) != 1 {
		t.Fatalf("len(Git.RecentCommits) = %d, want 1", len(result.Git.RecentCommits))
	}
}

func TestOverviewWorksWithoutGit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Project\n")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Overview(context.Background(), OverviewArgs{
		RootID: "repo",
	})
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}

	if result.Git.Available {
		t.Fatal("Git.Available = true, want false")
	}
	if result.Git.Error == "" {
		t.Fatal("Git.Error is empty")
	}
	if len(result.Warnings) == 0 {
		t.Fatal("Warnings is empty")
	}
}

func TestOverviewHonorsMaxEntries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "a")
	writeFile(t, dir, "b.txt", "b")
	writeFile(t, dir, "c.txt", "c")

	svc := newTestService(t, testRootConfig("repo", dir))

	result, err := svc.Overview(context.Background(), OverviewArgs{
		RootID:     "repo",
		MaxEntries: 2,
	})
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}

	if result.MaxEntries != 2 {
		t.Fatalf("MaxEntries = %d, want 2", result.MaxEntries)
	}
	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
}

func TestOverviewUsesInjectedRegistry(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cloudberry.project", "custom\n")
	writeFile(t, dir, "main.cloudberry", "source\n")
	writeFile(t, dir, "main.cloudberry.test", "test\n")

	registry := Registry{
		Project: ProjectRules{
			ImportantFiles:          []string{"Cloudberry.project"},
			SourceExtensions:        []string{".cloudberry"},
			TestPatterns:            []string{"*.cloudberry.test"},
			DocumentationExtensions: []string{},
			DocumentationFiles:      []string{},
			ConfigurationExtensions: []string{},
			ConfigurationFiles:      []string{"Cloudberry.project"},
		},
	}

	roots := makeTestRoots(t, testRootConfig("repo", dir))
	fsSvc := fsservice.New(roots, discardLogger())
	gitSvc := gitservice.New(roots, discardLogger())
	svc := NewWithRegistry(fsSvc, gitSvc, registry, "", discardLogger())

	result, err := svc.Overview(context.Background(), OverviewArgs{
		RootID:     "repo",
		MaxDepth:   1,
		MaxEntries: 20,
	})
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}

	if !containsString(result.ImportantFiles, "Cloudberry.project") {
		t.Fatalf("ImportantFiles = %#v, want Cloudberry.project", result.ImportantFiles)
	}
	if result.Counts.SourceFiles != 1 {
		t.Fatalf("SourceFiles = %d, want 1", result.Counts.SourceFiles)
	}
	if result.Counts.TestFiles != 1 {
		t.Fatalf("TestFiles = %d, want 1", result.Counts.TestFiles)
	}
	if result.Counts.ConfigurationFiles != 1 {
		t.Fatalf("ConfigurationFiles = %d, want 1", result.Counts.ConfigurationFiles)
	}
}

func newTestService(t *testing.T, configs ...config.RootConfig) *Service {
	t.Helper()

	roots := makeTestRoots(t, configs...)
	fsSvc := fsservice.New(roots, discardLogger())
	gitSvc := gitservice.New(roots, discardLogger())

	return NewWithRegistry(fsSvc, gitSvc, MustDefaultRegistryForTests(), "", discardLogger())
}

func makeTestRoots(t *testing.T, configs ...config.RootConfig) []*core.Root {
	t.Helper()

	roots := make([]*core.Root, 0, len(configs))
	for _, cfg := range configs {
		root, err := core.NewRoot(cfg, discardLogger())
		if err != nil {
			t.Fatalf("NewRoot(%q) returned error: %v", cfg.ID, err)
		}
		roots = append(roots, root)
	}

	return roots
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

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

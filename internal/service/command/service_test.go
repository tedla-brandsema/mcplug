package command

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
)

func TestCommandHelperProcess(t *testing.T) {
	if os.Getenv("MCPFS_COMMAND_HELPER") != "1" {
		return
	}

	args := os.Args
	for i, arg := range args {
		if arg == "--" && i+1 < len(args) {
			switch args[i+1] {
			case "ok":
				_, _ = os.Stdout.WriteString("helper stdout\n")
				_, _ = os.Stderr.WriteString("helper stderr\n")
				os.Exit(0)
			case "fail":
				_, _ = os.Stderr.WriteString("helper failed\n")
				os.Exit(7)
			case "cwd":
				cwd, err := os.Getwd()
				if err != nil {
					_, _ = os.Stderr.WriteString(err.Error())
					os.Exit(1)
				}
				_, _ = os.Stdout.WriteString(cwd)
				os.Exit(0)
			}
		}
	}

	os.Exit(2)
}

func TestListReturnsConfiguredCommands(t *testing.T) {
	dir := t.TempDir()
	svc := newTestCommandService(t, config.CommandConfig{
		Mode: config.CommandModePredefined,
		Items: []config.CommandItem{
			{
				ID:      "test",
				RootID:  "repo",
				Command: []string{"go", "test", "./..."},
			},
		},
	}, dir)

	result, err := svc.List(context.Background(), ListArgs{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if result.Mode != string(config.CommandModePredefined) {
		t.Fatalf("Mode = %q, want %q", result.Mode, config.CommandModePredefined)
	}
	if len(result.Commands) != 1 {
		t.Fatalf("len(Commands) = %d, want 1", len(result.Commands))
	}
	if result.Commands[0].ID != "test" {
		t.Fatalf("Command ID = %q, want test", result.Commands[0].ID)
	}
	if result.Commands[0].TimeoutSeconds != defaultTimeoutSeconds {
		t.Fatalf("TimeoutSeconds = %d, want %d", result.Commands[0].TimeoutSeconds, defaultTimeoutSeconds)
	}
}

func TestRunExecutesPredefinedCommand(t *testing.T) {
	t.Setenv("MCPFS_COMMAND_HELPER", "1")
	dir := t.TempDir()
	svc := newTestCommandService(t, config.CommandConfig{
		Mode: config.CommandModePredefined,
		Items: []config.CommandItem{
			{
				ID:      "ok",
				RootID:  "repo",
				Command: []string{os.Args[0], "-test.run=TestCommandHelperProcess", "--", "ok"},
			},
		},
	}, dir)

	result, err := svc.Run(context.Background(), RunArgs{ID: "ok"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0; stderr=%q", result.ExitCode, result.Stderr)
	}
	if result.Stdout != "helper stdout\n" {
		t.Fatalf("Stdout = %q, want helper stdout", result.Stdout)
	}
	if result.Stderr != "helper stderr\n" {
		t.Fatalf("Stderr = %q, want helper stderr", result.Stderr)
	}
	if result.TimedOut {
		t.Fatal("TimedOut = true, want false")
	}
}

func TestRunReturnsNonZeroExitCode(t *testing.T) {
	t.Setenv("MCPFS_COMMAND_HELPER", "1")
	dir := t.TempDir()
	svc := newTestCommandService(t, config.CommandConfig{
		Mode: config.CommandModePredefined,
		Items: []config.CommandItem{
			{
				ID:      "fail",
				RootID:  "repo",
				Command: []string{os.Args[0], "-test.run=TestCommandHelperProcess", "--", "fail"},
			},
		},
	}, dir)

	result, err := svc.Run(context.Background(), RunArgs{ID: "fail"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.ExitCode != 7 {
		t.Fatalf("ExitCode = %d, want 7", result.ExitCode)
	}
	if result.Stderr != "helper failed\n" {
		t.Fatalf("Stderr = %q, want helper failed", result.Stderr)
	}
}

func TestRunUsesRootScopedWorkdir(t *testing.T) {
	t.Setenv("MCPFS_COMMAND_HELPER", "1")
	dir := t.TempDir()
	workdir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatal(err)
	}

	svc := newTestCommandService(t, config.CommandConfig{
		Mode: config.CommandModePredefined,
		Items: []config.CommandItem{
			{
				ID:      "cwd",
				RootID:  "repo",
				Workdir: "subdir",
				Command: []string{os.Args[0], "-test.run=TestCommandHelperProcess", "--", "cwd"},
			},
		},
	}, dir)

	result, err := svc.Run(context.Background(), RunArgs{ID: "cwd"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Stdout != workdir {
		t.Fatalf("Stdout = %q, want %q", result.Stdout, workdir)
	}
}

func TestRunRejectsDisabledCommands(t *testing.T) {
	dir := t.TempDir()
	svc := newTestCommandService(t, config.CommandConfig{
		Mode: config.CommandModeDisabled,
		Items: []config.CommandItem{
			{
				ID:      "ok",
				RootID:  "repo",
				Command: []string{os.Args[0], "-test.run=TestCommandHelperProcess", "--", "ok"},
			},
		},
	}, dir)

	_, err := svc.Run(context.Background(), RunArgs{ID: "ok"})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	dir := t.TempDir()
	svc := newTestCommandService(t, config.CommandConfig{
		Mode: config.CommandModePredefined,
	}, dir)

	_, err := svc.Run(context.Background(), RunArgs{ID: "missing"})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
}

func TestNewRejectsEscapingWorkdir(t *testing.T) {
	dir := t.TempDir()
	root := newTestRoot(t, dir)

	_, err := New(config.CommandConfig{
		Mode: config.CommandModePredefined,
		Items: []config.CommandItem{
			{
				ID:      "bad",
				RootID:  "repo",
				Workdir: "..",
				Command: []string{"go", "test"},
			},
		},
	}, []*core.Root{root}, discardLogger())
	if err == nil {
		t.Fatal("New returned nil error")
	}
}

func newTestCommandService(t *testing.T, cfg config.CommandConfig, dir string) *Service {
	t.Helper()

	root := newTestRoot(t, dir)
	svc, err := New(cfg, []*core.Root{root}, discardLogger())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return svc
}

func newTestRoot(t *testing.T, dir string) *core.Root {
	t.Helper()

	root, err := core.NewRoot(config.RootConfig{
		ID:           "repo",
		Path:         dir,
		Mode:         config.ModeReadWrite,
		Include:      []string{"**/*"},
		UseGitignore: false,
		MaxFileBytes: 262144,
	}, discardLogger())
	if err != nil {
		t.Fatalf("NewRoot returned error: %v", err)
	}
	return root
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

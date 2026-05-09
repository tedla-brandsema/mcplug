package command

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

const (
	defaultTimeoutSeconds = 60
	defaultMaxOutputBytes = 65536
)

type Service struct {
	mode     config.CommandMode
	commands map[string]resolvedCommand
	order    []string
	logger   *slog.Logger
}

type resolvedCommand struct {
	id             string
	description    string
	rootID         string
	workdir        string
	workdirAbs     string
	command        []string
	timeoutSeconds int
	maxOutputBytes int
}

func New(cfg config.CommandConfig, roots []*core.Root, logger *slog.Logger) (*Service, error) {
	if logger == nil {
		logger = slog.Default()
	}

	rootsByID := make(map[string]*core.Root, len(roots))
	for _, root := range roots {
		rootsByID[root.ID] = root
	}

	svc := &Service{
		mode:     cfg.Mode,
		commands: make(map[string]resolvedCommand, len(cfg.Items)),
		order:    make([]string, 0, len(cfg.Items)),
		logger:   logger,
	}

	if svc.mode == "" {
		svc.mode = config.CommandModeDisabled
	}

	for _, item := range cfg.Items {
		root, ok := rootsByID[item.RootID]
		if !ok {
			return nil, fmt.Errorf("command %q references unknown root_id %q", item.ID, item.RootID)
		}

		workdir := item.Workdir
		if workdir == "" {
			workdir = "."
		}

		workdirAbs, err := core.ResolveInsideRoot(root.RealPath, workdir)
		if err != nil {
			return nil, fmt.Errorf("resolve command %q workdir: %w", item.ID, err)
		}

		workdirInfo, err := os.Stat(workdirAbs)
		if err != nil {
			return nil, fmt.Errorf("stat command %q workdir: %w", item.ID, err)
		}
		if !workdirInfo.IsDir() {
			return nil, fmt.Errorf("command %q workdir is not a directory", item.ID)
		}

		cmd := resolvedCommand{
			id:             item.ID,
			description:    item.Description,
			rootID:         item.RootID,
			workdir:        filepath.ToSlash(filepath.Clean(workdir)),
			workdirAbs:     workdirAbs,
			command:        append([]string(nil), item.Command...),
			timeoutSeconds: resolvePositiveInt(item.TimeoutSeconds, cfg.Defaults.TimeoutSeconds, defaultTimeoutSeconds),
			maxOutputBytes: resolvePositiveInt(item.MaxOutputBytes, cfg.Defaults.MaxOutputBytes, defaultMaxOutputBytes),
		}

		svc.commands[cmd.id] = cmd
		svc.order = append(svc.order, cmd.id)
	}

	sort.Strings(svc.order)
	return svc, nil
}

func (s *Service) Name() string {
	return "command"
}

func (s *Service) Mode() config.CommandMode {
	return s.mode
}

func (s *Service) Enabled() bool {
	return s.mode != config.CommandModeDisabled
}

func (s *Service) List(ctx context.Context, args ListArgs) (ListResult, error) {
	_ = ctx
	_ = args

	result := ListResult{
		Mode:     string(s.mode),
		Commands: make([]CommandInfo, 0, len(s.order)),
	}

	for _, id := range s.order {
		cmd := s.commands[id]
		result.Commands = append(result.Commands, CommandInfo{
			ID:             cmd.id,
			Description:    cmd.description,
			RootID:         cmd.rootID,
			Workdir:        cmd.workdir,
			Command:        append([]string(nil), cmd.command...),
			TimeoutSeconds: cmd.timeoutSeconds,
			MaxOutputBytes: cmd.maxOutputBytes,
		})
	}

	return result, nil
}

func (s *Service) Run(ctx context.Context, args RunArgs) (RunResult, error) {
	if s.mode == config.CommandModeDisabled {
		err := fmt.Errorf("command execution is disabled")
		s.logDenied(args.ID, err.Error())
		return RunResult{}, err
	}

	if args.ID == "" {
		err := fmt.Errorf("id is required")
		s.logDenied(args.ID, err.Error())
		return RunResult{}, err
	}

	cmd, ok := s.commands[args.ID]
	if !ok {
		err := fmt.Errorf("unknown command id %q", args.ID)
		s.logDenied(args.ID, err.Error())
		return RunResult{}, err
	}

	timeoutSeconds := resolvePositiveInt(args.TimeoutSeconds, cmd.timeoutSeconds, defaultTimeoutSeconds)
	maxOutputBytes := resolvePositiveInt(args.MaxOutputBytes, cmd.maxOutputBytes, defaultMaxOutputBytes)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()
	execCmd := exec.CommandContext(ctx, cmd.command[0], cmd.command[1:]...)
	execCmd.Dir = cmd.workdirAbs

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()
	duration := time.Since(start)
	timedOut := ctx.Err() == context.DeadlineExceeded
	exitCode := 0

	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		if timedOut {
			exitCode = -1
		}
	}

	stdoutText, stdoutTruncated := limits.CapStringBytes(stdout.String(), maxOutputBytes)
	remaining := maxOutputBytes - len(stdoutText)
	stderrText, stderrTruncated := limits.CapStringBytes(stderr.String(), remaining)

	result := RunResult{
		ID:             cmd.id,
		RootID:         cmd.rootID,
		Workdir:        cmd.workdir,
		Command:        append([]string(nil), cmd.command...),
		ExitCode:       exitCode,
		DurationMS:     duration.Milliseconds(),
		TimeoutSeconds: timeoutSeconds,
		MaxOutputBytes: maxOutputBytes,
		Stdout:         stdoutText,
		Stderr:         stderrText,
		Truncated:      stdoutTruncated || stderrTruncated,
		TimedOut:       timedOut,
	}

	s.logAllowed(cmd.id, "exit_code", result.ExitCode, "duration_ms", result.DurationMS, "timed_out", result.TimedOut)
	return result, nil
}

func resolvePositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func (s *Service) logAllowed(commandID string, attrs ...any) {
	args := []any{
		"service", s.Name(),
		"event", "mcpfs.command.run",
		"command_id", commandID,
	}
	args = append(args, attrs...)
	s.logger.Info("mcpfs allowed", args...)
}

func (s *Service) logDenied(commandID string, reason string) {
	s.logger.Warn(
		"mcpfs denied",
		slog.String("service", s.Name()),
		slog.String("event", "mcpfs.command.run"),
		slog.String("command_id", commandID),
		slog.String("reason", reason),
	)
}

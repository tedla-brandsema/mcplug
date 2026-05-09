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
	roots    map[string]*core.Root
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

type executionRequest struct {
	command        []string
	workdirAbs     string
	timeoutSeconds int
	maxOutputBytes int
}

type executionResult struct {
	exitCode       int
	durationMS     int64
	timeoutSeconds int
	maxOutputBytes int
	stdout         string
	stderr         string
	truncated      bool
	timedOut       bool
}

func New(cfg config.CommandConfig, roots []*core.Root, logger *slog.Logger) (*Service, error) {
	if logger == nil {
		logger = slog.Default()
	}

	svc := &Service{
		mode:     cfg.Mode,
		roots:    make(map[string]*core.Root, len(roots)),
		commands: make(map[string]resolvedCommand, len(cfg.Items)),
		order:    make([]string, 0, len(cfg.Items)),
		logger:   logger,
	}

	if svc.mode == "" {
		svc.mode = config.CommandModeDisabled
	}

	for _, root := range roots {
		svc.roots[root.ID] = root
	}

	for _, item := range cfg.Items {
		root, ok := svc.roots[item.RootID]
		if !ok {
			return nil, fmt.Errorf("command %q references unknown root_id %q", item.ID, item.RootID)
		}

		workdir := item.Workdir
		if workdir == "" {
			workdir = "."
		}

		workdirAbs, err := resolveCommandWorkdir(root, workdir)
		if err != nil {
			return nil, fmt.Errorf("resolve command %q workdir: %w", item.ID, err)
		}

		cmd := resolvedCommand{
			id:             item.ID,
			description:    item.Description,
			rootID:         item.RootID,
			workdir:        cleanCommandWorkdir(workdir),
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

func (s *Service) Unguarded() bool {
	return s.mode == config.CommandModeUnguarded
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
		s.logDenied("cmd_run", args.ID, err.Error())
		return RunResult{}, err
	}

	if args.ID == "" {
		err := fmt.Errorf("id is required")
		s.logDenied("cmd_run", args.ID, err.Error())
		return RunResult{}, err
	}

	cmd, ok := s.commands[args.ID]
	if !ok {
		err := fmt.Errorf("unknown command id %q", args.ID)
		s.logDenied("cmd_run", args.ID, err.Error())
		return RunResult{}, err
	}

	execResult := s.execute(ctx, executionRequest{
		command:        cmd.command,
		workdirAbs:     cmd.workdirAbs,
		timeoutSeconds: resolvePositiveInt(args.TimeoutSeconds, cmd.timeoutSeconds, defaultTimeoutSeconds),
		maxOutputBytes: resolvePositiveInt(args.MaxOutputBytes, cmd.maxOutputBytes, defaultMaxOutputBytes),
	})

	result := RunResult{
		ID:             cmd.id,
		RootID:         cmd.rootID,
		Workdir:        cmd.workdir,
		Command:        append([]string(nil), cmd.command...),
		ExitCode:       execResult.exitCode,
		DurationMS:     execResult.durationMS,
		TimeoutSeconds: execResult.timeoutSeconds,
		MaxOutputBytes: execResult.maxOutputBytes,
		Stdout:         execResult.stdout,
		Stderr:         execResult.stderr,
		Truncated:      execResult.truncated,
		TimedOut:       execResult.timedOut,
	}

	s.logAllowed("cmd_run", cmd.id, "exit_code", result.ExitCode, "duration_ms", result.DurationMS, "timed_out", result.TimedOut)
	return result, nil
}

func (s *Service) Exec(ctx context.Context, args ExecArgs) (ExecResult, error) {
	if s.mode != config.CommandModeUnguarded {
		err := fmt.Errorf("unguarded command execution is disabled")
		s.logDenied("cmd_exec", "", err.Error())
		return ExecResult{}, err
	}

	if args.RootID == "" {
		err := fmt.Errorf("root_id is required")
		s.logDenied("cmd_exec", "", err.Error())
		return ExecResult{}, err
	}

	root, ok := s.roots[args.RootID]
	if !ok {
		err := fmt.Errorf("unknown root_id %q", args.RootID)
		s.logDenied("cmd_exec", "", err.Error())
		return ExecResult{}, err
	}

	if len(args.Command) == 0 {
		err := fmt.Errorf("command is required")
		s.logDenied("cmd_exec", "", err.Error())
		return ExecResult{}, err
	}
	for i, arg := range args.Command {
		if arg == "" {
			err := fmt.Errorf("command[%d] must not be empty", i)
			s.logDenied("cmd_exec", "", err.Error())
			return ExecResult{}, err
		}
	}

	workdir := args.Workdir
	if workdir == "" {
		workdir = "."
	}

	workdirAbs, err := resolveCommandWorkdir(root, workdir)
	if err != nil {
		s.logDenied("cmd_exec", "", err.Error())
		return ExecResult{}, err
	}

	execResult := s.execute(ctx, executionRequest{
		command:        args.Command,
		workdirAbs:     workdirAbs,
		timeoutSeconds: resolvePositiveInt(args.TimeoutSeconds, defaultTimeoutSeconds),
		maxOutputBytes: resolvePositiveInt(args.MaxOutputBytes, defaultMaxOutputBytes),
	})

	result := ExecResult{
		RootID:         args.RootID,
		Workdir:        cleanCommandWorkdir(workdir),
		Command:        append([]string(nil), args.Command...),
		ExitCode:       execResult.exitCode,
		DurationMS:     execResult.durationMS,
		TimeoutSeconds: execResult.timeoutSeconds,
		MaxOutputBytes: execResult.maxOutputBytes,
		Stdout:         execResult.stdout,
		Stderr:         execResult.stderr,
		Truncated:      execResult.truncated,
		TimedOut:       execResult.timedOut,
	}

	s.logAllowed("cmd_exec", "", "root_id", result.RootID, "exit_code", result.ExitCode, "duration_ms", result.DurationMS, "timed_out", result.TimedOut)
	return result, nil
}

func (s *Service) execute(ctx context.Context, req executionRequest) executionResult {
	timeoutSeconds := resolvePositiveInt(req.timeoutSeconds, defaultTimeoutSeconds)
	maxOutputBytes := resolvePositiveInt(req.maxOutputBytes, defaultMaxOutputBytes)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()
	execCmd := exec.CommandContext(ctx, req.command[0], req.command[1:]...)
	execCmd.Dir = req.workdirAbs

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

	return executionResult{
		exitCode:       exitCode,
		durationMS:     duration.Milliseconds(),
		timeoutSeconds: timeoutSeconds,
		maxOutputBytes: maxOutputBytes,
		stdout:         stdoutText,
		stderr:         stderrText,
		truncated:      stdoutTruncated || stderrTruncated,
		timedOut:       timedOut,
	}
}

func resolveCommandWorkdir(root *core.Root, workdir string) (string, error) {
	workdirAbs, err := core.ResolveInsideRoot(root.RealPath, workdir)
	if err != nil {
		return "", err
	}

	workdirInfo, err := os.Stat(workdirAbs)
	if err != nil {
		return "", fmt.Errorf("stat workdir: %w", err)
	}
	if !workdirInfo.IsDir() {
		return "", fmt.Errorf("workdir is not a directory")
	}

	return workdirAbs, nil
}

func cleanCommandWorkdir(workdir string) string {
	return filepath.ToSlash(filepath.Clean(workdir))
}

func resolvePositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func (s *Service) logAllowed(tool string, commandID string, attrs ...any) {
	args := []any{
		"service", s.Name(),
		"event", "mcpfs.command.run",
		"tool", tool,
	}
	if commandID != "" {
		args = append(args, "command_id", commandID)
	}
	args = append(args, attrs...)
	s.logger.Info("mcpfs allowed", args...)
}

func (s *Service) logDenied(tool string, commandID string, reason string) {
	args := []any{
		"service", s.Name(),
		"event", "mcpfs.command.run",
		"tool", tool,
		"reason", reason,
	}
	if commandID != "" {
		args = append(args, "command_id", commandID)
	}
	s.logger.Warn("mcpfs denied", args...)
}

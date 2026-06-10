package upstream

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

// State is a supervisor lifecycle state.
type State string

const (
	StateStarting   State = "starting"
	StateRunning    State = "running"
	StateRestarting State = "restarting"
	StateFailed     State = "failed"
	StateStopping   State = "stopping"
	StateStopped    State = "stopped"
)

const (
	backoffMin        = 500 * time.Millisecond
	backoffMax        = 30 * time.Second
	backoffFactor     = 2
	healthyResetAfter = 60 * time.Second
)

// supervisor wraps a stdio upstream in an explicit lifecycle state machine:
// it spawns the child, watches for unexpected death, and respawns with
// backoff. Close never triggers a restart. While not running, CallTool
// returns tool errors per the Upstream contract. `optional` semantics live in
// StartAll: an initial Connect failure is simply returned; once Connect has
// succeeded the supervisor restarts the child regardless of optionality.
type supervisor struct {
	name   string
	cfg    config.MCPServer
	logger *slog.Logger

	bo           *backoff.Backoff
	healthyAfter time.Duration
	sleep        func(time.Duration)  // injectable for tests
	onTransition func(from, to State) // test hook, called outside mu

	mu        sync.Mutex
	state     State
	current   *stdioUpstream
	restarts  int
	startedAt time.Time
}

func newSupervisor(name string, cfg config.MCPServer, logger *slog.Logger) *supervisor {
	return &supervisor{
		name:   name,
		cfg:    cfg,
		logger: logger,
		bo: &backoff.Backoff{
			Min:    backoffMin,
			Max:    backoffMax,
			Factor: backoffFactor,
		},
		healthyAfter: healthyResetAfter,
		sleep:        time.Sleep,
		state:        StateStarting,
	}
}

func (s *supervisor) Name() string { return s.name }

// transition must be called with s.mu held.
func (s *supervisor) transition(to State, attrs ...any) {
	from := s.state
	s.state = to

	fields := append([]any{"upstream", s.name, "from", from, "to", to, "restarts", s.restarts}, attrs...)
	s.logger.Info("upstream state", fields...)

	if s.onTransition != nil {
		go s.onTransition(from, to)
	}
}

func (s *supervisor) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.state {
	case StateStarting, StateFailed:
	case StateRunning:
		return nil
	default:
		return fmt.Errorf("upstream %q is %s", s.name, s.state)
	}

	child := newStdioUpstream(s.name, s.cfg, s.logger)
	if err := child.Connect(ctx); err != nil {
		s.transition(StateFailed, "error", err)
		return err
	}

	s.current = child
	s.startedAt = time.Now()
	s.transition(StateRunning)
	go s.watch(child)
	return nil
}

// watch blocks until the given child exits, then decides between shutdown
// and restart based on the current state.
func (s *supervisor) watch(child *stdioUpstream) {
	err := child.Wait()

	s.mu.Lock()
	if s.state != StateRunning || s.current != child {
		// Close (or a newer generation) won the race; nothing to do.
		s.mu.Unlock()
		return
	}

	if time.Since(s.startedAt) > s.healthyAfter {
		s.bo.Reset()
	}
	s.current = nil
	s.transition(StateRestarting, "error", err)
	s.mu.Unlock()

	_ = child.Close()
	s.restartLoop()
}

func (s *supervisor) restartLoop() {
	for {
		delay := s.bo.Duration()
		s.logger.Info("upstream restart scheduled", "upstream", s.name, "delay", delay)
		s.sleep(delay)

		s.mu.Lock()
		if s.state != StateRestarting {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		child := newStdioUpstream(s.name, s.cfg, s.logger)
		if err := child.Connect(context.Background()); err != nil {
			_ = child.Close()
			s.logger.Warn("upstream restart failed", "upstream", s.name, "error", err)
			continue
		}

		s.mu.Lock()
		if s.state != StateRestarting {
			// Closed while we were respawning.
			s.mu.Unlock()
			_ = child.Close()
			return
		}
		s.current = child
		s.restarts++
		s.startedAt = time.Now()
		s.transition(StateRunning)
		s.mu.Unlock()

		go s.watch(child)
		return
	}
}

func (s *supervisor) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	child, state := s.snapshot()
	if child == nil {
		return nil, fmt.Errorf("upstream %q is %s", s.name, state)
	}
	return child.ListTools(ctx)
}

func (s *supervisor) CallTool(ctx context.Context, tool string, args map[string]any) (*mcp.CallToolResult, error) {
	child, state := s.snapshot()
	if child == nil {
		switch state {
		case StateRestarting:
			return toolError("upstream %q restarting", s.name), nil
		default:
			return toolError("upstream %q unavailable (%s)", s.name, state), nil
		}
	}
	return child.CallTool(ctx, tool, args)
}

func (s *supervisor) Close() error {
	s.mu.Lock()

	switch s.state {
	case StateStopping, StateStopped:
		s.mu.Unlock()
		return nil
	}

	s.transition(StateStopping)
	child := s.current
	s.current = nil
	s.mu.Unlock()

	var err error
	if child != nil {
		err = child.Close()
	}

	s.mu.Lock()
	s.transition(StateStopped)
	s.mu.Unlock()
	return err
}

func (s *supervisor) snapshot() (*stdioUpstream, State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != StateRunning || s.current == nil {
		return nil, s.state
	}
	return s.current, s.state
}

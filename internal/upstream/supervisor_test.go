package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

const childEnvMarker = "MCPFS_TEST_CHILD"

// TestMain doubles as the test child: when the marker env var is set, the
// test binary serves a tiny MCP server over stdio instead of running tests.
func TestMain(m *testing.M) {
	if os.Getenv(childEnvMarker) == "1" {
		runChildServer()
		return
	}
	os.Exit(m.Run())
}

type pidResult struct {
	PID int `json:"pid"`
}

func runChildServer() {
	fmt.Fprintln(os.Stderr, "child ready")

	server := mcp.NewServer(&mcp.Implementation{Name: "test-child", Version: "test"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pid",
		Description: "returns the child process id",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, pidResult, error) {
		return nil, pidResult{PID: os.Getpid()}, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, "child server:", err)
		os.Exit(1)
	}
}

func childConfig(t *testing.T) config.MCPServer {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	return config.MCPServer{
		Command: exe,
		Env:     map[string]string{childEnvMarker: "1"},
	}
}

// transitionLog records supervisor state transitions race-safely.
type transitionLog struct {
	mu     sync.Mutex
	states []State
}

func (l *transitionLog) record(_, to State) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.states = append(l.states, to)
}

func (l *transitionLog) has(s State) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, got := range l.states {
		if got == s {
			return true
		}
	}
	return false
}

func newTestSupervisor(t *testing.T, logger *slog.Logger) (*supervisor, *transitionLog) {
	t.Helper()
	if logger == nil {
		logger = testLogger(t)
	}

	s := newSupervisor("child", childConfig(t), logger)
	s.bo.Min = 50 * time.Millisecond
	s.bo.Max = 400 * time.Millisecond
	s.healthyAfter = time.Hour // never reset unless a test opts in

	log := &transitionLog{}
	s.onTransition = log.record

	t.Cleanup(func() { _ = s.Close() })
	return s, log
}

func childPID(t *testing.T, s *supervisor) int {
	t.Helper()

	result, err := s.CallTool(t.Context(), "pid", nil)
	if err != nil {
		t.Fatalf("CallTool(pid) returned protocol error: %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool(pid) returned tool error: %s", resultText(result))
	}

	var parsed pidResult
	if err := json.Unmarshal([]byte(resultText(result)), &parsed); err != nil {
		t.Fatalf("parse pid result %q: %v", resultText(result), err)
	}
	return parsed.PID
}

// waitForNewPID polls until the supervisor serves calls from a process other
// than oldPID.
func waitForNewPID(t *testing.T, s *supervisor, oldPID int) int {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		result, err := s.CallTool(t.Context(), "pid", nil)
		if err != nil {
			t.Fatalf("CallTool(pid) returned protocol error: %v", err)
		}
		if !result.IsError {
			var parsed pidResult
			if err := json.Unmarshal([]byte(resultText(result)), &parsed); err == nil && parsed.PID != oldPID {
				return parsed.PID
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("supervisor did not recover with a new child in time")
	return 0
}

func TestSupervisorRestartsAfterUnexpectedExit(t *testing.T) {
	s, log := newTestSupervisor(t, nil)
	if err := s.Connect(t.Context()); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	oldPID := childPID(t, s)
	if err := syscall.Kill(oldPID, syscall.SIGKILL); err != nil {
		t.Fatalf("kill child: %v", err)
	}

	newPID := waitForNewPID(t, s, oldPID)
	if newPID == oldPID {
		t.Fatal("child pid did not change after restart")
	}
	if !log.has(StateRestarting) {
		t.Fatal("no restarting transition recorded")
	}
}

func TestSupervisorCloseDoesNotRestart(t *testing.T) {
	s, log := newTestSupervisor(t, nil)
	if err := s.Connect(t.Context()); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	time.Sleep(300 * time.Millisecond) // give a buggy watcher time to misbehave

	if log.has(StateRestarting) {
		t.Fatal("Close triggered a restart")
	}
	if _, state := s.snapshot(); state != StateStopped {
		t.Fatalf("state = %s, want %s", state, StateStopped)
	}

	result, err := s.CallTool(t.Context(), "pid", nil)
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("CallTool after Close should return a tool error")
	}
}

func TestSupervisorCallDuringRestartReturnsToolError(t *testing.T) {
	s, log := newTestSupervisor(t, nil)
	s.bo.Min = 2 * time.Second // keep it restarting long enough to observe
	s.bo.Max = 2 * time.Second

	if err := s.Connect(t.Context()); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	pid := childPID(t, s)
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatalf("kill child: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for !log.has(StateRestarting) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !log.has(StateRestarting) {
		t.Fatal("restarting transition not observed")
	}

	result, err := s.CallTool(t.Context(), "pid", nil)
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v (want tool error)", err)
	}
	if !result.IsError {
		t.Fatal("CallTool during restart should return a tool error")
	}
	if !strings.Contains(resultText(result), "restarting") {
		t.Fatalf("tool error = %q, want mention of restarting", resultText(result))
	}
}

func TestSupervisorBackoffGrowsAcrossCrashes(t *testing.T) {
	s, _ := newTestSupervisor(t, nil)

	var mu sync.Mutex
	var delays []time.Duration
	s.sleep = func(d time.Duration) {
		mu.Lock()
		delays = append(delays, d)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}

	if err := s.Connect(t.Context()); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	pid := childPID(t, s)
	for range 2 {
		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
			t.Fatalf("kill child: %v", err)
		}
		pid = waitForNewPID(t, s, pid)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(delays) < 2 {
		t.Fatalf("recorded %d restart delays, want >= 2", len(delays))
	}
	if delays[1] <= delays[0] {
		t.Fatalf("backoff did not grow: %v then %v", delays[0], delays[1])
	}
}

func TestSupervisorBackoffResetsAfterHealthyPeriod(t *testing.T) {
	s, _ := newTestSupervisor(t, nil)
	s.healthyAfter = 0 // every run counts as healthy

	var mu sync.Mutex
	var delays []time.Duration
	s.sleep = func(d time.Duration) {
		mu.Lock()
		delays = append(delays, d)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}

	if err := s.Connect(t.Context()); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	pid := childPID(t, s)
	for range 2 {
		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
			t.Fatalf("kill child: %v", err)
		}
		pid = waitForNewPID(t, s, pid)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(delays) < 2 {
		t.Fatalf("recorded %d restart delays, want >= 2", len(delays))
	}
	if delays[0] != delays[1] {
		t.Fatalf("backoff did not reset after healthy period: %v then %v", delays[0], delays[1])
	}
}

func TestSupervisorStartupFailureReturnsError(t *testing.T) {
	s := newSupervisor("broken", config.MCPServer{Command: "/no/such/binary/anywhere"}, testLogger(t))
	t.Cleanup(func() { _ = s.Close() })

	if err := s.Connect(t.Context()); err == nil {
		t.Fatal("Connect returned nil error for missing binary")
	}
	if _, state := s.snapshot(); state != StateFailed {
		t.Fatalf("state = %s, want %s", state, StateFailed)
	}
}

func TestChildStderrIsLoggedWithUpstreamName(t *testing.T) {
	var buf syncBuffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	s, _ := newTestSupervisor(t, logger)
	if err := s.Connect(t.Context()); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		out := buf.String()
		if strings.Contains(out, "child ready") && strings.Contains(out, "upstream=child") {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("child stderr not attributed in logs:\n%s", buf.String())
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

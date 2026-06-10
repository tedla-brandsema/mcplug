package upstream

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sort"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

// stdioUpstream runs an MCP server as a persistent child process and reuses
// one client session across calls. Restart supervision is layered on top by
// the supervisor; this type itself is single-shot.
type stdioUpstream struct {
	name   string
	cfg    config.MCPServer
	logger *slog.Logger

	mu      sync.Mutex
	session *mcp.ClientSession
	closed  bool
}

func newStdioUpstream(name string, cfg config.MCPServer, logger *slog.Logger) *stdioUpstream {
	return &stdioUpstream{
		name:   name,
		cfg:    cfg,
		logger: logger,
	}
}

func (u *stdioUpstream) Name() string { return u.name }

// newCmd builds a fresh, unstarted command. The SDK's CommandTransport owns
// Start and shutdown escalation; an exec.Cmd cannot be reused, so every
// (re)connect needs a new one. The command runs verbatim, never via a shell.
func (u *stdioUpstream) newCmd(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, u.cfg.Command, u.cfg.Args...)
	cmd.Dir = u.cfg.Cwd
	cmd.Env = mergeEnv(os.Environ(), u.cfg.Env)
	cmd.Stderr = newLineLogger(u.logger, u.name)
	return cmd
}

func (u *stdioUpstream) Connect(ctx context.Context) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.closed {
		return fmt.Errorf("upstream %q is closed", u.name)
	}
	if u.session != nil {
		return nil
	}

	// The child must outlive the connect timeout, so the command context is
	// detached from ctx on purpose.
	session, err := mcp.NewClient(clientImplementation(), nil).Connect(ctx, &mcp.CommandTransport{
		Command: u.newCmd(context.Background()),
	}, nil)
	if err != nil {
		return fmt.Errorf("spawn upstream %q: %w", u.name, err)
	}

	u.session = session
	return nil
}

func (u *stdioUpstream) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	session, err := u.currentSession()
	if err != nil {
		return nil, err
	}

	ctx, cancel := withTimeout(ctx)
	defer cancel()

	result, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
}

func (u *stdioUpstream) CallTool(ctx context.Context, tool string, args map[string]any) (*mcp.CallToolResult, error) {
	session, err := u.currentSession()
	if err != nil {
		return toolError("upstream %q unavailable: %v", u.name, err), nil
	}

	ctx, cancel := withTimeout(ctx)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      tool,
		Arguments: args,
	})
	if err != nil {
		return callFailure(u.name, "call failed", err), nil
	}
	return result, nil
}

func (u *stdioUpstream) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.closed = true
	if u.session == nil {
		return nil
	}

	session := u.session
	u.session = nil
	return session.Close()
}

// Wait blocks until the child process exits. Used by the supervisor to
// detect unexpected death.
func (u *stdioUpstream) Wait() error {
	session, err := u.currentSession()
	if err != nil {
		return err
	}
	return session.Wait()
}

func (u *stdioUpstream) currentSession() (*mcp.ClientSession, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.closed {
		return nil, fmt.Errorf("upstream %q is closed", u.name)
	}
	if u.session == nil {
		return nil, fmt.Errorf("upstream %q is not connected", u.name)
	}
	return u.session, nil
}

// mergeEnv appends config env entries to the inherited environment in sorted
// key order; for duplicate keys the config value wins because exec uses the
// last occurrence.
func mergeEnv(base []string, extra map[string]string) []string {
	if len(extra) == 0 {
		return base
	}

	keys := make([]string, 0, len(extra))
	for k := range extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	merged := make([]string, 0, len(base)+len(extra))
	merged = append(merged, base...)
	for _, k := range keys {
		merged = append(merged, k+"="+extra[k])
	}
	return merged
}

// newLineLogger returns a writer that logs each child stderr line attributed
// to its upstream. Values are logged as-is: stderr is child-controlled output,
// not config secrets.
func newLineLogger(logger *slog.Logger, name string) io.Writer {
	pr, pw := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			logger.Info("upstream stderr", "upstream", name, "line", scanner.Text())
		}
	}()
	return pw
}

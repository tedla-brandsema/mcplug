package fs

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
)

type Service struct {
	roots  map[string]*core.Root
	order  []string
	logger *slog.Logger
}

func New(roots []*core.Root, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Service{
		roots:  make(map[string]*core.Root, len(roots)),
		order:  make([]string, 0, len(roots)),
		logger: logger,
	}

	for _, root := range roots {
		s.roots[root.ID] = root
		s.order = append(s.order, root.ID)
	}

	sort.Strings(s.order)
	return s
}

// NewService keeps the old construction shape available while the app is being
// refactored. Later, server assembly can create roots once and call New directly.
func NewService(cfg config.Config, logger *slog.Logger) (*Service, error) {
	if logger == nil {
		logger = slog.Default()
	}

	roots := make([]*core.Root, 0, len(cfg.Roots))
	for _, rootCfg := range cfg.Roots {
		root, err := core.NewRoot(rootCfg, logger)
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}

	return New(roots, logger), nil
}

func (s *Service) Name() string {
	return "fs"
}

func (s *Service) Roots(ctx context.Context, args RootsArgs) (RootsResult, error) {
	_ = ctx
	_ = args

	out := RootsResult{
		Roots: make([]RootInfo, 0, len(s.order)),
	}

	for _, id := range s.order {
		root := s.roots[id]
		out.Roots = append(out.Roots, RootInfo{
			ID:           root.ID,
			Mode:         string(root.Mode),
			MaxFileBytes: root.MaxFileBytes,
		})
	}

	s.logger.Info("mcpfs allowed", "service", s.Name(), "event", "mcpfs.roots", "roots", len(out.Roots))
	return out, nil
}

func (s *Service) root(id string) (*core.Root, error) {
	if id == "" {
		return nil, fmt.Errorf("root_id is required")
	}

	root, ok := s.roots[id]
	if !ok {
		return nil, fmt.Errorf("unknown root_id %q", id)
	}

	return root, nil
}

func (s *Service) resolve(root *core.Root, requested string) (string, error) {
	abs, err := core.ResolveInsideRoot(root.RealPath, requested)
	if err != nil {
		return "", err
	}

	rel, err := root.Rel(abs)
	if err != nil {
		return "", err
	}

	return cleanFSRel(rel), nil
}

func (s *Service) logAllowed(event string, rootID string, path string, attrs ...any) {
	args := []any{
		"service", s.Name(),
		"event", event,
		"root_id", rootID,
		"path", path,
	}
	args = append(args, attrs...)

	s.logger.Info("mcpfs allowed", args...)
}

func (s *Service) logDenied(event string, rootID string, path string, reason string) {
	s.logger.Warn(
		"mcpfs denied",
		slog.String("service", s.Name()),
		slog.String("event", event),
		slog.String("root_id", rootID),
		slog.String("path", path),
		slog.String("reason", reason),
	)
}

func joinRel(base string, name string) string {
	base = cleanFSRel(base)
	name = path.Clean(filepath.ToSlash(name))

	if base == "." {
		return name
	}

	return path.Join(base, name)
}

func cleanFSRel(rel string) string {
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "" || rel == "." {
		return "."
	}
	return strings.TrimPrefix(rel, "./")
}
package git

import (
	"fmt"
	"log/slog"
	"sort"

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

func (s *Service) Name() string {
	return "git"
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

func (s *Service) resolve(root *core.Root, requested string) (string, error) {
	abs, err := core.ResolveInsideRoot(root.RealPath, requested)
	if err != nil {
		return "", err
	}

	rel, err := root.Rel(abs)
	if err != nil {
		return "", err
	}

	return rel, nil
}
package project

import (
	"fmt"
	"log/slog"

	fsservice "github.com/tedla-brandsema/mcpfs/internal/service/fs"
	gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"
)

type Service struct {
	fs           *fsservice.Service
	git          *gitservice.Service
	registry     Registry
	registryPath string
	logger       *slog.Logger
}

func New(fsSvc *fsservice.Service, gitSvc *gitservice.Service, logger *slog.Logger) (*Service, error) {
	registry, registryPath, err := LoadOrCreateDefaultRegistry()
	if err != nil {
		return nil, fmt.Errorf("load project registry: %w", err)
	}

	return NewWithRegistry(fsSvc, gitSvc, registry, registryPath, logger), nil
}

func NewWithRegistry(fsSvc *fsservice.Service, gitSvc *gitservice.Service, registry Registry, registryPath string, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		fs:           fsSvc,
		git:          gitSvc,
		registry:     registry,
		registryPath: registryPath,
		logger:       logger,
	}
}

func (s *Service) Name() string {
	return "project"
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

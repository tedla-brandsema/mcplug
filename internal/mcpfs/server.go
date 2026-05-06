package mcpfs

import (
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
	fsservice "github.com/tedla-brandsema/mcpfs/internal/service/fs"
	gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"
	projectservice "github.com/tedla-brandsema/mcpfs/internal/service/project"
)

type Server struct {
	MCP     *mcp.Server
	FS      *fsservice.Service
	Git     *gitservice.Service
	Project *projectservice.Service
}

func NewServer(cfg config.Config, logger *slog.Logger) (*Server, error) {
	if logger == nil {
		logger = slog.Default()
	}

	roots := make([]*core.Root, 0, len(cfg.Roots))
	for _, rootCfg := range cfg.Roots {
		root, err := core.NewRoot(rootCfg, logger)
		if err != nil {
			return nil, fmt.Errorf("create root %q: %w", rootCfg.ID, err)
		}
		roots = append(roots, root)
	}

	fsSvc := fsservice.New(roots, logger)
	gitSvc := gitservice.New(roots, logger)
	projectSvc, err := projectservice.New(fsSvc, gitSvc, logger)
	if err != nil {
		return nil, err
	}

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)

	RegisterFSTools(mcpServer, fsSvc)
	RegisterGitTools(mcpServer, gitSvc)
	RegisterProjectTools(mcpServer, projectSvc)

	return &Server{
		MCP:     mcpServer,
		FS:      fsSvc,
		Git:     gitSvc,
		Project: projectSvc,
	}, nil
}

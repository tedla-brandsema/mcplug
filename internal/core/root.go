package core

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

const defaultMaxFileBytes int64 = 262144

type Root struct {
	ID           string
	Path         string
	RealPath     string
	Mode         config.Mode
	MaxFileBytes int64
	Matcher      *Matcher
	ReadFS       fs.FS
}

func NewRoot(cfg config.RootConfig, logger *slog.Logger) (*Root, error) {
	abs, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("abs root path %q: %w", cfg.Path, err)
	}

	realPath, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("resolve root path %q: %w", abs, err)
	}

	info, err := os.Stat(realPath)
	if err != nil {
		return nil, fmt.Errorf("stat root path %q: %w", realPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path %q is not a directory", realPath)
	}

	maxFileBytes := cfg.MaxFileBytes
	if maxFileBytes == 0 {
		maxFileBytes = defaultMaxFileBytes
	}

	matcher, err := NewMatcher(realPath, cfg.Include, cfg.Exclude, cfg.UseGitignore, logger)
	if err != nil {
		return nil, fmt.Errorf("create matcher for root %q: %w", cfg.ID, err)
	}

return &Root{
	ID:           cfg.ID,
	Path:         abs,
	RealPath:     realPath,
	Mode:         cfg.Mode,
	MaxFileBytes: maxFileBytes,
	Matcher:      matcher,
	ReadFS:       os.DirFS(realPath),
}, nil
}

func (r *Root) Rel(abs string) (string, error) {
	rel, err := filepath.Rel(r.RealPath, abs)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return ".", nil
	}
	return filepath.ToSlash(rel), nil
}
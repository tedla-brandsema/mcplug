package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
)

func (s *Service) Write(ctx context.Context, args WriteArgs) (WriteResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.write", args.RootID, args.Path, err.Error())
		return WriteResult{}, err
	}

	if root.Mode != config.ModeReadWrite {
		err := fmt.Errorf("root %q is not writable", root.ID)
		s.logDenied("mcpfs.write", root.ID, args.Path, err.Error())
		return WriteResult{}, err
	}

	if int64(len(args.Content)) > root.MaxFileBytes {
		err := fmt.Errorf("content exceeds max_file_bytes: size=%d max=%d", len(args.Content), root.MaxFileBytes)
		s.logDenied("mcpfs.write", root.ID, args.Path, err.Error())
		return WriteResult{}, err
	}

	abs, err := core.ResolveWritableInsideRoot(root.RealPath, args.Path)
	if err != nil {
		s.logDenied("mcpfs.write", root.ID, args.Path, err.Error())
		return WriteResult{}, err
	}

	rel, err := root.Rel(abs)
	if err != nil {
		s.logDenied("mcpfs.write", root.ID, args.Path, err.Error())
		return WriteResult{}, err
	}
	rel = cleanFSRel(rel)

	if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.write", root.ID, rel, err.Error())
		return WriteResult{}, err
	}

	info, err := os.Stat(abs)
	if err == nil && info.IsDir() {
		err := fmt.Errorf("path is a directory")
		s.logDenied("mcpfs.write", root.ID, rel, err.Error())
		return WriteResult{}, err
	}
	if err != nil && !os.IsNotExist(err) {
		s.logDenied("mcpfs.write", root.ID, rel, err.Error())
		return WriteResult{}, err
	}

	if args.ExpectedSHA256 != "" {
		if err != nil {
			s.logDenied("mcpfs.write", root.ID, rel, "expected_sha256 requires an existing file")
			return WriteResult{}, fmt.Errorf("expected_sha256 requires an existing file")
		}

		if _, err := s.verifyExpectedSHA256(root, rel, args.ExpectedSHA256); err != nil {
			s.logDenied("mcpfs.write", root.ID, rel, err.Error())
			return WriteResult{}, err
		}
	}

	parent := filepath.Dir(abs)
	parentInfo, err := os.Stat(parent)
	if err != nil {
		if !os.IsNotExist(err) {
			s.logDenied("mcpfs.write", root.ID, rel, err.Error())
			return WriteResult{}, err
		}

		if !args.CreateDirs {
			err := fmt.Errorf("parent directory does not exist")
			s.logDenied("mcpfs.write", root.ID, rel, err.Error())
			return WriteResult{}, err
		}

		if err := os.MkdirAll(parent, 0o755); err != nil {
			s.logDenied("mcpfs.write", root.ID, rel, err.Error())
			return WriteResult{}, err
		}
	} else if !parentInfo.IsDir() {
		err := fmt.Errorf("parent path is not a directory")
		s.logDenied("mcpfs.write", root.ID, rel, err.Error())
		return WriteResult{}, err
	}

	if err := os.WriteFile(abs, []byte(args.Content), fs.FileMode(0o644)); err != nil {
		s.logDenied("mcpfs.write", root.ID, rel, err.Error())
		return WriteResult{}, err
	}

	result := WriteResult{
		RootID: root.ID,
		Path:   rel,
		Bytes:  len(args.Content),
		Mode:   string(root.Mode),
	}

	s.logAllowed("mcpfs.write", root.ID, rel, "bytes", result.Bytes)
	return result, nil
}

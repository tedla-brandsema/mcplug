package fs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/tedla-brandsema/mcpfs/internal/core"
)

func (s *Service) Hash(ctx context.Context, args HashArgs) (HashResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.hash", args.RootID, args.Path, err.Error())
		return HashResult{}, err
	}

	result, err := s.hashFile(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.hash", root.ID, args.Path, err.Error())
		return HashResult{}, err
	}

	s.logAllowed("mcpfs.hash", root.ID, result.Path, "bytes", result.Size)
	return result, nil
}

func (s *Service) hashFile(root *core.Root, requested string) (HashResult, error) {
	abs, err := core.ResolveInsideRoot(root.RealPath, requested)
	if err != nil {
		return HashResult{}, err
	}

	rel, err := root.Rel(abs)
	if err != nil {
		return HashResult{}, err
	}
	rel = cleanFSRel(rel)

	if !root.Matcher.AllowFile(rel) {
		return HashResult{}, fmt.Errorf("file is excluded")
	}

	info, err := os.Stat(abs)
	if err != nil {
		return HashResult{}, err
	}
	if info.IsDir() {
		return HashResult{}, fmt.Errorf("path is a directory")
	}
	if info.Size() > root.MaxFileBytes {
		return HashResult{}, fmt.Errorf("file exceeds max_file_bytes: size=%d max=%d", info.Size(), root.MaxFileBytes)
	}

	file, err := os.Open(abs)
	if err != nil {
		return HashResult{}, err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return HashResult{}, err
	}

	return HashResult{
		RootID: root.ID,
		Path:   rel,
		SHA256: hex.EncodeToString(h.Sum(nil)),
		Size:   info.Size(),
		MTime:  info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
		Mode:   info.Mode().String(),
	}, nil
}

func (s *Service) verifyExpectedSHA256(root *core.Root, requested string, expected string) (HashResult, error) {
	if expected == "" {
		return HashResult{}, nil
	}

	actual, err := s.hashFile(root, requested)
	if err != nil {
		return HashResult{}, err
	}

	if actual.SHA256 != expected {
		return HashResult{}, fmt.Errorf("sha256 mismatch")
	}

	return actual, nil
}

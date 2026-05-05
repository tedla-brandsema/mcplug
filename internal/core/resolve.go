package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveInsideRoot(rootPath string, requested string) (string, error) {
	if requested == "" {
		requested = "."
	}

	if filepath.IsAbs(requested) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	clean := filepath.Clean(requested)
	if clean == "." {
		clean = "."
	}

	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root")
	}

	rootReal, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	candidate := filepath.Join(rootReal, clean)

	if realCandidate, err := filepath.EvalSymlinks(candidate); err == nil {
		candidate = realCandidate
	}

	rel, err := filepath.Rel(rootReal, candidate)
	if err != nil {
		return "", fmt.Errorf("relative path check: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("resolved path escapes root")
	}

	return candidate, nil
}
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

	if err := requireInsideRoot(rootReal, candidate); err != nil {
		return "", err
	}

	return candidate, nil
}

func ResolveWritableInsideRoot(rootPath string, requested string) (string, error) {
	if requested == "" {
		return "", fmt.Errorf("path is required")
	}

	if filepath.IsAbs(requested) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	clean := filepath.Clean(requested)
	if clean == "." {
		return "", fmt.Errorf("path must reference a file")
	}

	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root")
	}

	rootReal, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	parentRel := filepath.Dir(clean)
	base := filepath.Base(clean)

	parentAbs := filepath.Join(rootReal, parentRel)
	parentReal, err := resolveWritableParent(rootReal, parentAbs)
	if err != nil {
		return "", err
	}

	candidate := filepath.Join(parentReal, base)
	if err := requireInsideRoot(rootReal, candidate); err != nil {
		return "", err
	}

	return candidate, nil
}

func resolveWritableParent(rootReal string, parentAbs string) (string, error) {
	current := parentAbs
	missing := []string{}

	for {
		real, err := filepath.EvalSymlinks(current)
		if err == nil {
			if err := requireInsideRoot(rootReal, real); err != nil {
				return "", err
			}

			for i := len(missing) - 1; i >= 0; i-- {
				real = filepath.Join(real, missing[i])
			}

			if err := requireInsideRoot(rootReal, real); err != nil {
				return "", err
			}

			return real, nil
		}

		if !os.IsNotExist(err) {
			return "", fmt.Errorf("resolve writable parent %q: %w", current, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("resolve writable parent %q: no existing parent", parentAbs)
		}

		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

func requireInsideRoot(rootReal string, candidate string) error {
	rel, err := filepath.Rel(rootReal, candidate)
	if err != nil {
		return fmt.Errorf("relative path check: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("resolved path escapes root")
	}

	return nil
}

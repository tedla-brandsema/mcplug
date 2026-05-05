package git

import (
	"context"
	"fmt"
)

func (s *Service) Diff(ctx context.Context, args DiffArgs) (DiffResult, error) {
	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("git.diff", args.RootID, args.Path, err.Error())
		return DiffResult{}, err
	}

	maxBytes := args.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultGitOutputLimit
	}
	if maxBytes > defaultGitOutputLimit {
		maxBytes = defaultGitOutputLimit
	}

	gitArgs := []string{"diff"}
	if args.Staged {
		gitArgs = append(gitArgs, "--cached")
	}

	pathForResult := ""
	if args.Path != "" {
		rel, err := s.resolve(root, args.Path)
		if err != nil {
			s.logDenied("git.diff", root.ID, args.Path, err.Error())
			return DiffResult{}, err
		}

		pathForResult = rel
		gitArgs = append(gitArgs, "--", rel)
	}

	stdout, stderr, truncated, err := runGit(ctx, root.RealPath, maxBytes, gitArgs...)
	if err != nil {
		err := fmt.Errorf("git diff: %w: %s", err, stderr)
		s.logDenied("git.diff", root.ID, pathForResult, err.Error())
		return DiffResult{}, err
	}

	result := DiffResult{
		RootID:    root.ID,
		Path:      pathForResult,
		Staged:    args.Staged,
		Bytes:     len(stdout),
		Truncated: truncated,
		Diff:      stdout,
	}

	s.logAllowed(
		"git.diff",
		root.ID,
		pathForResult,
		"staged", result.Staged,
		"bytes", result.Bytes,
		"truncated", result.Truncated,
	)

	return result, nil
}
package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	iofs "io/fs"
	"strings"

	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

func (s *Service) Diff(ctx context.Context, args DiffArgs) (DiffResult, error) {
	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("git.diff", args.RootID, args.Path, err.Error())
		return DiffResult{}, err
	}

	maxBytes := limits.ClampInt(args.MaxBytes, defaultGitOutputLimit, defaultGitOutputLimit)

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

		if !root.Matcher.AllowFile(rel) {
			err := fmt.Errorf("file is excluded")
			s.logDenied("git.diff", root.ID, rel, err.Error())
			return DiffResult{}, err
		}

		pathForResult = rel

		if !args.Staged {
			untracked, err := s.isUntracked(ctx, root, rel)
			if err != nil {
				s.logDenied("git.diff", root.ID, rel, err.Error())
				return DiffResult{}, err
			}

			if untracked {
				result, err := s.diffUntracked(ctx, root, rel, maxBytes)
				if err != nil {
					s.logDenied("git.diff", root.ID, rel, err.Error())
					return DiffResult{}, err
				}

				s.logAllowed(
					"git.diff",
					root.ID,
					rel,
					"staged", result.Staged,
					"bytes", result.Bytes,
					"truncated", result.Truncated,
					"synthetic", true,
				)

				return result, nil
			}
		}

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
		"synthetic", false,
	)

	return result, nil
}

func (s *Service) isUntracked(ctx context.Context, root *core.Root, rel string) (bool, error) {
	stdout, stderr, _, err := runGit(ctx, root.RealPath, defaultGitOutputLimit,
		"status",
		"--porcelain=v1",
		"--untracked-files=all",
		"--",
		rel,
	)
	if err != nil {
		return false, fmt.Errorf("git status for path: %w: %s", err, stderr)
	}

	_, changes, err := ParseStatus(stdout)
	if err != nil {
		return false, err
	}

	for _, change := range changes {
		if change.Path == rel && change.Status == "untracked" {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) diffUntracked(ctx context.Context, root *core.Root, rel string, maxBytes int) (DiffResult, error) {
	_ = ctx

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		return DiffResult{}, err
	}
	if info.IsDir() {
		return DiffResult{}, fmt.Errorf("path is a directory")
	}
	if info.Size() > root.MaxFileBytes {
		return DiffResult{}, fmt.Errorf("file exceeds max_file_bytes: size=%d max=%d", info.Size(), root.MaxFileBytes)
	}

	f, err := root.ReadFS.Open(rel)
	if err != nil {
		return DiffResult{}, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return DiffResult{}, err
	}

	diff, truncated := syntheticNewFileDiff(rel, data, info.Mode(), maxBytes)

	return DiffResult{
		RootID:    root.ID,
		Path:      rel,
		Staged:    false,
		Bytes:     len(diff),
		Truncated: truncated,
		Diff:      diff,
	}, nil
}

func syntheticNewFileDiff(rel string, data []byte, mode iofs.FileMode, maxBytes int) (string, bool) {
	maxBytes = limits.ClampInt(maxBytes, defaultGitOutputLimit, defaultGitOutputLimit)

	fileMode := "100644"
	if mode.Perm()&0o111 != 0 {
		fileMode = "100755"
	}

	lineCount := countDiffLines(data)

	var out cappedBuffer
	out.limit = maxBytes

	fmt.Fprintf(&out, "diff --git a/%s b/%s\n", rel, rel)
	fmt.Fprintf(&out, "new file mode %s\n", fileMode)
	fmt.Fprintf(&out, "index 0000000..0000000\n")
	fmt.Fprintf(&out, "--- /dev/null\n")
	fmt.Fprintf(&out, "+++ b/%s\n", rel)
	fmt.Fprintf(&out, "@@ -0,0 +1,%d @@\n", lineCount)

	content := string(data)
	if content == "" {
		return out.String(), out.truncated
	}

	lines := strings.SplitAfter(content, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fmt.Fprintf(&out, "+%s", line)
		if out.truncated {
			return out.String(), true
		}
	}

	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		fmt.Fprintf(&out, "\n\\ No newline at end of file\n")
	}

	return out.String(), out.truncated
}

func countDiffLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	count := bytes.Count(data, []byte("\n"))
	if !bytes.HasSuffix(data, []byte("\n")) {
		count++
	}

	return count
}

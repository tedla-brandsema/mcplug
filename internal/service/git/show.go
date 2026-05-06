package git

import (
	"context"
	"fmt"
	"strings"
)

const showCommitFormat = "%H%x00%h%x00%an%x00%ae%x00%ad%x00%s%x00%b"

func (s *Service) Show(ctx context.Context, args ShowArgs) (ShowResult, error) {
	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("git.show", args.RootID, args.Path, err.Error())
		return ShowResult{}, err
	}

	rev, err := validateShowRev(args.Rev)
	if err != nil {
		s.logDenied("git.show", root.ID, args.Path, err.Error())
		return ShowResult{}, err
	}

	maxBytes := args.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultGitOutputLimit
	}
	if maxBytes > defaultGitOutputLimit {
		maxBytes = defaultGitOutputLimit
	}

	commitHash, err := s.resolveCommit(ctx, root.RealPath, rev)
	if err != nil {
		s.logDenied("git.show", root.ID, args.Path, err.Error())
		return ShowResult{}, err
	}

	commit, err := s.showCommitMetadata(ctx, root.RealPath, commitHash)
	if err != nil {
		s.logDenied("git.show", root.ID, args.Path, err.Error())
		return ShowResult{}, err
	}

	pathForResult := ""
	gitArgs := []string{
		"show",
		"--no-ext-diff",
		"--patch",
		"--pretty=format:",
		commitHash,
	}

	if args.Path != "" {
		rel, err := s.resolve(root, args.Path)
		if err != nil {
			s.logDenied("git.show", root.ID, args.Path, err.Error())
			return ShowResult{}, err
		}

		if !root.Matcher.AllowFile(rel) && !root.Matcher.AllowDir(rel) {
			err := fmt.Errorf("path is excluded")
			s.logDenied("git.show", root.ID, rel, err.Error())
			return ShowResult{}, err
		}

		pathForResult = rel
		gitArgs = append(gitArgs, "--", rel)
	}

	stdout, stderr, truncated, err := runGit(ctx, root.RealPath, maxBytes, gitArgs...)
	if err != nil {
		err := fmt.Errorf("git show: %w: %s", err, stderr)
		s.logDenied("git.show", root.ID, pathForResult, err.Error())
		return ShowResult{}, err
	}

	diff := strings.TrimLeft(stdout, "\r\n")
	diff, diffTruncated := capStringBytes(diff, maxBytes)
	truncated = truncated || diffTruncated

	result := ShowResult{
		RootID:    root.ID,
		Rev:       rev,
		Path:      pathForResult,
		Commit:    commit,
		Bytes:     len(diff),
		Truncated: truncated,
		Diff:      diff,
	}

	s.logAllowed(
		"git.show",
		root.ID,
		pathForResult,
		"rev", rev,
		"commit", commit.Hash,
		"bytes", result.Bytes,
		"truncated", result.Truncated,
	)

	return result, nil
}

func validateShowRev(rev string) (string, error) {
	if rev == "" {
		return "", fmt.Errorf("rev is required")
	}

	clean := strings.TrimSpace(rev)
	if clean == "" {
		return "", fmt.Errorf("rev is required")
	}
	if clean != rev {
		return "", fmt.Errorf("rev must not contain leading or trailing whitespace")
	}
	if strings.HasPrefix(clean, "-") {
		return "", fmt.Errorf("rev must not start with '-'")
	}
	if strings.Contains(clean, "\x00") {
		return "", fmt.Errorf("rev must not contain NUL bytes")
	}

	return clean, nil
}

func (s *Service) resolveCommit(ctx context.Context, repoPath string, rev string) (string, error) {
	stdout, stderr, _, err := runGit(ctx, repoPath, 1024,
		"rev-parse",
		"--verify",
		"--quiet",
		rev+"^{commit}",
	)
	if err != nil {
		return "", fmt.Errorf("resolve rev %q: %w: %s", rev, err, stderr)
	}

	commitHash := strings.TrimSpace(stdout)
	if commitHash == "" {
		return "", fmt.Errorf("resolve rev %q: empty commit hash", rev)
	}

	return commitHash, nil
}

func (s *Service) showCommitMetadata(ctx context.Context, repoPath string, commitHash string) (ShowCommit, error) {
	stdout, stderr, _, err := runGit(ctx, repoPath, 64*1024,
		"show",
		"--no-patch",
		"--date=iso-strict",
		"--pretty=format:"+showCommitFormat,
		commitHash,
	)
	if err != nil {
		return ShowCommit{}, fmt.Errorf("git show metadata: %w: %s", err, stderr)
	}

	commit, err := ParseShowCommit(stdout)
	if err != nil {
		return ShowCommit{}, err
	}

	return commit, nil
}

func capStringBytes(s string, maxBytes int) (string, bool) {
	if maxBytes <= 0 {
		return "", len(s) > 0
	}
	if len(s) <= maxBytes {
		return s, false
	}
	return s[:maxBytes], true
}

func ParseShowCommit(output string) (ShowCommit, error) {
	fields := strings.Split(output, "\x00")
	if len(fields) < 7 {
		return ShowCommit{}, fmt.Errorf("invalid git show metadata: expected at least 7 fields, got %d", len(fields))
	}

	return ShowCommit{
		Hash:        fields[0],
		ShortHash:   fields[1],
		AuthorName:  fields[2],
		AuthorEmail: fields[3],
		AuthorDate:  fields[4],
		Subject:     fields[5],
		Body:        strings.TrimSpace(fields[6]),
	}, nil
}

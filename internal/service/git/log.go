package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

const (
	defaultGitLogLimit = 10
	maxGitLogLimit     = 100
)

func (s *Service) Log(ctx context.Context, args LogArgs) (LogResult, error) {
	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("git.log", args.RootID, args.Path, err.Error())
		return LogResult{}, err
	}

	limit := limits.ClampInt(args.Limit, defaultGitLogLimit, maxGitLogLimit)
	maxBytes := limits.ClampInt(args.MaxBytes, defaultGitOutputLimit, defaultGitOutputLimit)

	gitArgs := []string{
		"log",
		"--date=iso-strict",
		"--pretty=format:%H%x00%h%x00%an%x00%ae%x00%ad%x00%s%x00%b%x00%x1e",
		"-n",
		strconv.Itoa(limit),
	}

	pathForResult := ""
	if args.Path != "" {
		rel, err := s.resolve(root, args.Path)
		if err != nil {
			s.logDenied("git.log", root.ID, args.Path, err.Error())
			return LogResult{}, err
		}

		pathForResult = rel
		gitArgs = append(gitArgs, "--", rel)
	}

	stdout, stderr, truncated, err := runGit(ctx, root.RealPath, maxBytes, gitArgs...)
	if err != nil {
		err := fmt.Errorf("git log: %w: %s", err, stderr)
		s.logDenied("git.log", root.ID, pathForResult, err.Error())
		return LogResult{}, err
	}

	commits, err := ParseLog(stdout)
	if err != nil {
		s.logDenied("git.log", root.ID, pathForResult, err.Error())
		return LogResult{}, err
	}

	result := LogResult{
		RootID:    root.ID,
		Path:      pathForResult,
		Limit:     limit,
		Commits:   commits,
		Truncated: truncated,
	}

	s.logAllowed(
		"git.log",
		root.ID,
		pathForResult,
		"limit", limit,
		"commits", len(commits),
		"truncated", truncated,
	)

	return result, nil
}

func ParseLog(output string) ([]LogCommit, error) {
	output = strings.TrimSuffix(output, "\x1e")
	if strings.TrimSpace(output) == "" {
		return []LogCommit{}, nil
	}

	records := strings.Split(output, "\x1e")
	commits := make([]LogCommit, 0, len(records))

	for _, record := range records {
		record = strings.Trim(record, "\n\r")
		if record == "" {
			continue
		}

		fields := strings.Split(record, "\x00")
		if len(fields) < 7 {
			return nil, fmt.Errorf("invalid git log record: expected at least 7 fields, got %d", len(fields))
		}

		commit := LogCommit{
			Hash:        fields[0],
			ShortHash:   fields[1],
			AuthorName:  fields[2],
			AuthorEmail: fields[3],
			AuthorDate:  fields[4],
			Subject:     fields[5],
			Body:        strings.TrimSpace(fields[6]),
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

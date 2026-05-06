package fs

import (
	"bufio"
	"context"
	"fmt"
	iofs "io/fs"

	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

const (
	defaultReadLinesWindow = 200
	maxReadLinesWindow     = 1000
)

func (s *Service) ReadLines(ctx context.Context, args ReadLinesArgs) (ReadLinesResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.read_lines", args.RootID, args.Path, err.Error())
		return ReadLinesResult{}, err
	}

	rel, err := s.resolve(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.read_lines", root.ID, args.Path, err.Error())
		return ReadLinesResult{}, err
	}

	if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.read_lines", root.ID, rel, err.Error())
		return ReadLinesResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.read_lines", root.ID, rel, err.Error())
		return ReadLinesResult{}, err
	}
	if info.IsDir() {
		err := fmt.Errorf("path is a directory")
		s.logDenied("mcpfs.read_lines", root.ID, rel, err.Error())
		return ReadLinesResult{}, err
	}
	if info.Size() > root.MaxFileBytes {
		err := fmt.Errorf("file exceeds max_file_bytes: size=%d max=%d", info.Size(), root.MaxFileBytes)
		s.logDenied("mcpfs.read_lines", root.ID, rel, err.Error())
		return ReadLinesResult{}, err
	}

	startLine := args.StartLine
	if startLine <= 0 {
		startLine = 1
	}

	endLine := args.EndLine
	if endLine <= 0 {
		endLine = startLine + defaultReadLinesWindow - 1
	}
	if endLine < startLine {
		err := fmt.Errorf("end_line must be >= start_line")
		s.logDenied("mcpfs.read_lines", root.ID, rel, err.Error())
		return ReadLinesResult{}, err
	}

	window := endLine - startLine + 1
	window = limits.ClampInt(window, defaultReadLinesWindow, maxReadLinesWindow)
	endLine = startLine + window - 1

	f, err := root.ReadFS.Open(rel)
	if err != nil {
		s.logDenied("mcpfs.read_lines", root.ID, rel, err.Error())
		return ReadLinesResult{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	result := ReadLinesResult{
		RootID:    root.ID,
		Path:      rel,
		StartLine: startLine,
		EndLine:   endLine,
		Lines:     make([]ReadLine, 0),
	}

	lineNo := 0
	for scanner.Scan() {
		lineNo++

		if lineNo < startLine {
			continue
		}

		if lineNo > endLine {
			result.Truncated = true
			break
		}

		result.Lines = append(result.Lines, ReadLine{
			Number: lineNo,
			Text:   scanner.Text(),
		})
	}

	if err := scanner.Err(); err != nil {
		s.logDenied("mcpfs.read_lines", root.ID, rel, err.Error())
		return ReadLinesResult{}, err
	}

	s.logAllowed(
		"mcpfs.read_lines",
		root.ID,
		rel,
		"start_line", result.StartLine,
		"end_line", result.EndLine,
		"lines", len(result.Lines),
		"truncated", result.Truncated,
	)

	return result, nil
}

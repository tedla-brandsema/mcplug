package fs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	iofs "io/fs"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/tedla-brandsema/mcpfs/internal/core"
)

func (s *Service) List(ctx context.Context, args ListArgs) (ListResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.list", args.RootID, args.Path, err.Error())
		return ListResult{}, err
	}

	requested := args.Path
	if requested == "" {
		requested = "."
	}

	rel, err := s.resolve(root, requested)
	if err != nil {
		s.logDenied("mcpfs.list", root.ID, requested, err.Error())
		return ListResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}
	if !info.IsDir() {
		err := fmt.Errorf("path is not a directory")
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}
	if !root.Matcher.AllowDir(rel) {
		err := fmt.Errorf("directory is excluded")
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}

	maxEntries := args.MaxEntries
	if maxEntries <= 0 {
		maxEntries = 200
	}
	if maxEntries > 1000 {
		maxEntries = 1000
	}

	result := ListResult{
		RootID:  root.ID,
		Path:    rel,
		Entries: make([]Entry, 0),
	}

	if args.Recursive {
		err = iofs.WalkDir(root.ReadFS, rel, func(pathRel string, d iofs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}

			if pathRel == rel {
				return nil
			}

			safeRel, err := s.resolve(root, pathRel)
			if err != nil {
				s.logDenied("mcpfs.list", root.ID, pathRel, err.Error())
				if d.IsDir() {
					return iofs.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				if !root.Matcher.AllowDir(safeRel) {
					return iofs.SkipDir
				}
			} else if !root.Matcher.AllowFile(safeRel) {
				return nil
			}

			if len(result.Entries) >= maxEntries {
				result.Truncated = true
				if d.IsDir() {
					return iofs.SkipDir
				}
				return nil
			}

			entry, err := makeEntry(root, safeRel)
			if err == nil {
				result.Entries = append(result.Entries, entry)
			}

			return nil
		})
	} else {
		var entries []iofs.DirEntry
		entries, err = iofs.ReadDir(root.ReadFS, rel)
		if err == nil {
			for _, d := range entries {
				if len(result.Entries) >= maxEntries {
					result.Truncated = true
					break
				}

				entryRel := joinRel(rel, d.Name())

				safeRel, err := s.resolve(root, entryRel)
				if err != nil {
					s.logDenied("mcpfs.list", root.ID, entryRel, err.Error())
					continue
				}

				if d.IsDir() {
					if !root.Matcher.AllowDir(safeRel) {
						continue
					}
				} else if !root.Matcher.AllowFile(safeRel) {
					continue
				}

				entry, err := makeEntry(root, safeRel)
				if err == nil {
					result.Entries = append(result.Entries, entry)
				}
			}
		}
	}

	if err != nil {
		s.logDenied("mcpfs.list", root.ID, rel, err.Error())
		return ListResult{}, err
	}

	s.logAllowed("mcpfs.list", root.ID, rel, "entries", len(result.Entries), "truncated", result.Truncated)
	return result, nil
}

func (s *Service) Read(ctx context.Context, args ReadArgs) (ReadResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.read", args.RootID, args.Path, err.Error())
		return ReadResult{}, err
	}

	rel, err := s.resolve(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, args.Path, err.Error())
		return ReadResult{}, err
	}

	if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	if info.IsDir() {
		err := fmt.Errorf("path is a directory")
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	if info.Size() > root.MaxFileBytes {
		err := fmt.Errorf("file exceeds max_file_bytes: size=%d max=%d", info.Size(), root.MaxFileBytes)
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	if args.Offset < 0 {
		err := fmt.Errorf("offset must be >= 0")
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}

	limit := args.Limit
	if limit <= 0 || limit > root.MaxFileBytes {
		limit = root.MaxFileBytes
	}

	f, err := root.ReadFS.Open(rel)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	defer f.Close()

	if args.Offset > 0 {
		n, err := io.CopyN(io.Discard, f, args.Offset)
		if err != nil && err != io.EOF {
			s.logDenied("mcpfs.read", root.ID, rel, err.Error())
			return ReadResult{}, err
		}
		if n < args.Offset {
			result := ReadResult{
				RootID:    root.ID,
				Path:      rel,
				Bytes:     0,
				Size:      info.Size(),
				Offset:    args.Offset,
				Truncated: false,
				Content:   "",
			}
			s.logAllowed("mcpfs.read", root.ID, rel, "bytes", result.Bytes, "truncated", result.Truncated)
			return result, nil
		}
	}

	data, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}

	result := ReadResult{
		RootID:    root.ID,
		Path:      rel,
		Bytes:     len(data),
		Size:      info.Size(),
		Offset:    args.Offset,
		Truncated: args.Offset+int64(len(data)) < info.Size(),
		Content:   string(data),
	}

	s.logAllowed("mcpfs.read", root.ID, rel, "bytes", result.Bytes, "truncated", result.Truncated)
	return result, nil
}

func (s *Service) Search(ctx context.Context, args SearchArgs) (SearchResult, error) {
	_ = ctx

	if args.Query == "" {
		return SearchResult{}, fmt.Errorf("query is required")
	}

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.search", args.RootID, "", err.Error())
		return SearchResult{}, err
	}

	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = 50
	}
	if maxResults > 500 {
		maxResults = 500
	}

	glob := filepath.ToSlash(strings.TrimSpace(args.Glob))

	result := SearchResult{
		RootID:  root.ID,
		Query:   args.Query,
		Matches: make([]SearchMatch, 0),
	}

	err = iofs.WalkDir(root.ReadFS, ".", func(pathRel string, d iofs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if pathRel == "." {
			return nil
		}

		safeRel, err := s.resolve(root, pathRel)
		if err != nil {
			s.logDenied("mcpfs.search", root.ID, pathRel, err.Error())
			if d.IsDir() {
				return iofs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if !root.Matcher.AllowDir(safeRel) {
				return iofs.SkipDir
			}
			return nil
		}

		if !root.Matcher.AllowFile(safeRel) {
			return nil
		}

		if glob != "" {
			ok, _ := doublestar.PathMatch(glob, safeRel)
			if !ok {
				return nil
			}
		}

		info, err := iofs.Stat(root.ReadFS, safeRel)
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Size() > root.MaxFileBytes {
			return nil
		}

		matches, err := searchFile(root, safeRel, args.Query, maxResults-len(result.Matches))
		if err != nil {
			return nil
		}

		result.Matches = append(result.Matches, matches...)

		if len(result.Matches) >= maxResults {
			result.Truncated = true
			return iofs.SkipAll
		}

		return nil
	})
	if err != nil {
		s.logDenied("mcpfs.search", root.ID, args.Query, err.Error())
		return SearchResult{}, err
	}

	s.logAllowed("mcpfs.search", root.ID, args.Query, "matches", len(result.Matches), "truncated", result.Truncated)
	return result, nil
}

func (s *Service) Stat(ctx context.Context, args StatArgs) (StatResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.stat", args.RootID, args.Path, err.Error())
		return StatResult{}, err
	}

	rel, err := s.resolve(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.stat", root.ID, args.Path, err.Error())
		return StatResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.stat", root.ID, rel, err.Error())
		return StatResult{}, err
	}

	if info.IsDir() {
		if !root.Matcher.AllowDir(rel) {
			err := fmt.Errorf("directory is excluded")
			s.logDenied("mcpfs.stat", root.ID, rel, err.Error())
			return StatResult{}, err
		}
	} else if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.stat", root.ID, rel, err.Error())
		return StatResult{}, err
	}

	typ := "file"
	if info.IsDir() {
		typ = "dir"
	}

	result := StatResult{
		RootID: root.ID,
		Path:   rel,
		Type:   typ,
		Size:   info.Size(),
		MTime:  info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
		Mode:   info.Mode().String(),
	}

	s.logAllowed("mcpfs.stat", root.ID, rel, "type", typ)
	return result, nil
}

func makeEntry(root *core.Root, rel string) (Entry, error) {
	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		return Entry{}, err
	}

	typ := "file"
	if info.IsDir() {
		typ = "dir"
	}

	return Entry{
		Path:  filepath.ToSlash(rel),
		Type:  typ,
		Size:  info.Size(),
		MTime: info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

func searchFile(root *core.Root, rel string, query string, remaining int) ([]SearchMatch, error) {
	if remaining <= 0 {
		return nil, nil
	}

	f, err := root.ReadFS.Open(rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var matches []SearchMatch
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if strings.Contains(line, query) {
			matches = append(matches, SearchMatch{
				Path:    rel,
				Line:    lineNo,
				Preview: strings.TrimSpace(line),
			})

			if len(matches) >= remaining {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return matches, err
	}

	return matches, nil
}
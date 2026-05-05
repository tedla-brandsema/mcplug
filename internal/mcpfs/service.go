package mcpfs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

type Service struct {
	roots  map[string]*Root
	order  []string
	logger *slog.Logger
}

func NewService(cfg config.Config, logger *slog.Logger) (*Service, error) {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Service{
		roots:  make(map[string]*Root, len(cfg.Roots)),
		order:  make([]string, 0, len(cfg.Roots)),
		logger: logger,
	}

	for _, rootCfg := range cfg.Roots {
		root, err := NewRoot(rootCfg, logger)
		if err != nil {
			return nil, err
		}

		s.roots[root.ID] = root
		s.order = append(s.order, root.ID)
	}

	sort.Strings(s.order)
	return s, nil
}

type RootsArgs struct{}

type RootInfo struct {
	ID           string `json:"id"`
	Mode         string `json:"mode"`
	MaxFileBytes int64  `json:"max_file_bytes"`
}

type RootsResult struct {
	Roots []RootInfo `json:"roots"`
}

type ListArgs struct {
	RootID     string `json:"root_id" jsonschema:"configured root id"`
	Path       string `json:"path,omitempty" jsonschema:"relative directory path inside the root"`
	Recursive  bool   `json:"recursive,omitempty" jsonschema:"whether to list recursively"`
	MaxEntries int    `json:"max_entries,omitempty" jsonschema:"maximum number of entries to return"`
}

type Entry struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Size  int64  `json:"size,omitempty"`
	MTime string `json:"mtime,omitempty"`
}

type ListResult struct {
	RootID    string  `json:"root_id"`
	Path      string  `json:"path"`
	Entries   []Entry `json:"entries"`
	Truncated bool    `json:"truncated"`
}

type ReadArgs struct {
	RootID string `json:"root_id" jsonschema:"configured root id"`
	Path   string `json:"path" jsonschema:"relative file path inside the root"`
	Offset int64  `json:"offset,omitempty" jsonschema:"byte offset"`
	Limit  int64  `json:"limit,omitempty" jsonschema:"maximum bytes to read"`
}

type ReadResult struct {
	RootID    string `json:"root_id"`
	Path      string `json:"path"`
	Bytes     int    `json:"bytes"`
	Size      int64  `json:"size"`
	Offset    int64  `json:"offset"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
}

type SearchArgs struct {
	RootID     string `json:"root_id" jsonschema:"configured root id"`
	Query      string `json:"query" jsonschema:"case-sensitive substring query"`
	Glob       string `json:"glob,omitempty" jsonschema:"optional glob such as **/*.go"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"maximum number of matches"`
}

type SearchMatch struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Preview string `json:"preview"`
}

type SearchResult struct {
	RootID    string        `json:"root_id"`
	Query     string        `json:"query"`
	Matches   []SearchMatch `json:"matches"`
	Truncated bool          `json:"truncated"`
}

type StatArgs struct {
	RootID string `json:"root_id" jsonschema:"configured root id"`
	Path   string `json:"path" jsonschema:"relative path inside the root"`
}

type StatResult struct {
	RootID string `json:"root_id"`
	Path   string `json:"path"`
	Type   string `json:"type"`
	Size   int64  `json:"size,omitempty"`
	MTime  string `json:"mtime,omitempty"`
	Mode   string `json:"mode,omitempty"`
}

func (s *Service) Roots(ctx context.Context, args RootsArgs) (RootsResult, error) {
	_ = ctx

	out := RootsResult{
		Roots: make([]RootInfo, 0, len(s.order)),
	}

	for _, id := range s.order {
		root := s.roots[id]
		out.Roots = append(out.Roots, RootInfo{
			ID:           root.ID,
			Mode:         string(root.Mode),
			MaxFileBytes: root.MaxFileBytes,
		})
	}

	s.logger.Info("mcpfs allowed", "event", "mcpfs.roots", "roots", len(out.Roots))
	return out, nil
}

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

	abs, rel, err := s.resolve(root, requested)
	if err != nil {
		s.logDenied("mcpfs.list", root.ID, requested, err.Error())
		return ListResult{}, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		s.logDenied("mcpfs.list", root.ID, requested, err.Error())
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
		err = filepath.WalkDir(abs, func(pathAbs string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}

			if pathAbs == abs {
				return nil
			}

			entryRel, err := root.Rel(pathAbs)
			if err != nil {
				return nil
			}

			if d.IsDir() {
				if !root.Matcher.AllowDir(entryRel) {
					return filepath.SkipDir
				}
			} else if !root.Matcher.AllowFile(entryRel) {
				return nil
			}

			if len(result.Entries) >= maxEntries {
				result.Truncated = true
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			entry, err := makeEntry(entryRel, d)
			if err == nil {
				result.Entries = append(result.Entries, entry)
			}

			return nil
		})
	} else {
		var entries []os.DirEntry
		entries, err = os.ReadDir(abs)
		if err == nil {
			for _, d := range entries {
				if len(result.Entries) >= maxEntries {
					result.Truncated = true
					break
				}

				entryAbs := filepath.Join(abs, d.Name())
				entryRel, err := root.Rel(entryAbs)
				if err != nil {
					continue
				}

				if d.IsDir() {
					if !root.Matcher.AllowDir(entryRel) {
						continue
					}
				} else if !root.Matcher.AllowFile(entryRel) {
					continue
				}

				entry, err := makeEntry(entryRel, d)
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

	abs, rel, err := s.resolve(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, args.Path, err.Error())
		return ReadResult{}, err
	}

	if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}

	info, err := os.Stat(abs)
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

	f, err := os.Open(abs)
	if err != nil {
		s.logDenied("mcpfs.read", root.ID, rel, err.Error())
		return ReadResult{}, err
	}
	defer f.Close()

	if args.Offset > 0 {
		if _, err := f.Seek(args.Offset, io.SeekStart); err != nil {
			s.logDenied("mcpfs.read", root.ID, rel, err.Error())
			return ReadResult{}, err
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

	err = filepath.WalkDir(root.RealPath, func(abs string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if abs == root.RealPath {
			return nil
		}

		rel, err := root.Rel(abs)
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if !root.Matcher.AllowDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}

		if !root.Matcher.AllowFile(rel) {
			return nil
		}

		if glob != "" {
			ok, _ := doublestar.PathMatch(glob, rel)
			if !ok {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > root.MaxFileBytes {
			return nil
		}

		matches, err := searchFile(abs, rel, args.Query, maxResults-len(result.Matches))
		if err != nil {
			return nil
		}

		result.Matches = append(result.Matches, matches...)

		if len(result.Matches) >= maxResults {
			result.Truncated = true
			return filepath.SkipAll
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

	abs, rel, err := s.resolve(root, args.Path)
	if err != nil {
		s.logDenied("mcpfs.stat", root.ID, args.Path, err.Error())
		return StatResult{}, err
	}

	info, err := os.Stat(abs)
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

func (s *Service) root(id string) (*Root, error) {
	if id == "" {
		return nil, fmt.Errorf("root_id is required")
	}

	root, ok := s.roots[id]
	if !ok {
		return nil, fmt.Errorf("unknown root_id %q", id)
	}

	return root, nil
}

func (s *Service) resolve(root *Root, requested string) (string, string, error) {
	abs, err := ResolveInsideRoot(root.RealPath, requested)
	if err != nil {
		return "", "", err
	}

	rel, err := root.Rel(abs)
	if err != nil {
		return "", "", err
	}

	return abs, rel, nil
}

func makeEntry(rel string, d os.DirEntry) (Entry, error) {
	info, err := d.Info()
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

func searchFile(abs string, rel string, query string, remaining int) ([]SearchMatch, error) {
	if remaining <= 0 {
		return nil, nil
	}

	f, err := os.Open(abs)
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

func jsonString(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(data)
}
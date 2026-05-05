package core

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type Matcher struct {
	root       string
	include    []string
	exclude    []string
	gitignore  []gitignoreRule
	useGit     bool
	logger     *slog.Logger
	ruleSerial int
}

type gitignoreRule struct {
	Base    string
	Pattern string
	Negated bool
	DirOnly bool
	Order   int
}

func NewMatcher(root string, include []string, exclude []string, useGitignore bool, logger *slog.Logger) (*Matcher, error) {
	m := &Matcher{
		root:    root,
		include: normalizePatterns(include),
		exclude: normalizePatterns(exclude),
		useGit:  useGitignore,
		logger:  logger,
	}

	if useGitignore {
		if err := m.loadGitignoreRules(); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func normalizePatterns(patterns []string) []string {
	out := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		if pattern == "" {
			continue
		}
		out = append(out, pattern)
	}
	return out
}

func (m *Matcher) AllowFile(rel string) bool {
	rel = cleanRel(rel)
	if rel == "." {
		return true
	}

	if m.matchesAny(m.exclude, rel, false) {
		return false
	}

	if m.isGitIgnored(rel, false) {
		return false
	}

	if len(m.include) == 0 {
		return true
	}

	return m.matchesAny(m.include, rel, false)
}

func (m *Matcher) AllowDir(rel string) bool {
	rel = cleanRel(rel)
	if rel == "." {
		return true
	}

	if m.matchesAny(m.exclude, rel, true) {
		return false
	}

	if m.isGitIgnored(rel, true) {
		return false
	}

	// Include rules restrict files, not traversal. A directory may contain
	// included files further below it.
	return true
}

func (m *Matcher) matchesAny(patterns []string, rel string, isDir bool) bool {
	for _, pattern := range patterns {
		if matchPattern(pattern, rel, isDir) {
			return true
		}
	}
	return false
}

func matchPattern(pattern string, rel string, isDir bool) bool {
	pattern = strings.TrimSpace(filepath.ToSlash(pattern))
	rel = cleanRel(rel)

	if pattern == "" {
		return false
	}

	dirOnly := strings.HasSuffix(pattern, "/")
	if dirOnly {
		pattern = strings.TrimSuffix(pattern, "/")
		if !isDir && !pathHasPrefix(rel, pattern) {
			return false
		}
	}

	if ok, _ := doublestar.PathMatch(pattern, rel); ok {
		return true
	}

	if strings.HasPrefix(pattern, "**/") {
		if ok, _ := doublestar.PathMatch(strings.TrimPrefix(pattern, "**/"), rel); ok {
			return true
		}
	}

	if dirOnly && pathHasPrefix(rel, pattern) {
		return true
	}

	return false
}

func (m *Matcher) loadGitignoreRules() error {
	var rules []gitignoreRule

	err := filepath.WalkDir(m.root, func(abs string, d os.DirEntry, err error) error {
		if err != nil {
			if m.logger != nil {
				m.logger.Warn("walk gitignore", "path", abs, "error", err)
			}
			return nil
		}

		name := d.Name()
		if d.IsDir() && name == ".git" {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		if name != ".gitignore" {
			return nil
		}

		baseAbs := filepath.Dir(abs)
		baseRel, err := filepath.Rel(m.root, baseAbs)
		if err != nil {
			return nil
		}
		baseRel = cleanRel(baseRel)
		if baseRel == "." {
			baseRel = ""
		}

		fileRules, err := parseGitignoreFile(abs, baseRel, &m.ruleSerial)
		if err != nil {
			if m.logger != nil {
				m.logger.Warn("parse gitignore", "path", abs, "error", err)
			}
			return nil
		}

		rules = append(rules, fileRules...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk gitignore files: %w", err)
	}

	sort.SliceStable(rules, func(i, j int) bool {
		di := pathDepth(rules[i].Base)
		dj := pathDepth(rules[j].Base)
		if di != dj {
			return di < dj
		}
		return rules[i].Order < rules[j].Order
	})

	m.gitignore = rules
	return nil
}

func parseGitignoreFile(filename string, base string, serial *int) ([]gitignoreRule, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rules []gitignoreRule
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t\r")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, `\#`) {
			line = strings.TrimPrefix(line, `\`)
		}

		negated := false
		if strings.HasPrefix(line, "!") {
			negated = true
			line = strings.TrimPrefix(line, "!")
			}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		dirOnly := strings.HasSuffix(line, "/")
		line = strings.TrimSuffix(line, "/")
		line = filepath.ToSlash(line)

		*serial++

		rules = append(rules, gitignoreRule{
			Base:    base,
			Pattern: line,
			Negated: negated,
			DirOnly: dirOnly,
			Order:   *serial,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (m *Matcher) isGitIgnored(rel string, isDir bool) bool {
	if !m.useGit || len(m.gitignore) == 0 {
		return false
	}

	rel = cleanRel(rel)
	ignored := false

	for _, rule := range m.gitignore {
		if rule.matches(rel, isDir) {
			ignored = !rule.Negated
		}
	}

	return ignored
}

func (r gitignoreRule) matches(rel string, isDir bool) bool {
	rel = cleanRel(rel)

	if r.Base != "" {
		if rel != r.Base && !strings.HasPrefix(rel, r.Base+"/") {
			return false
		}
	}

	sub := rel
	if r.Base != "" {
		sub = strings.TrimPrefix(rel, r.Base)
		sub = strings.TrimPrefix(sub, "/")
	}
	if sub == "" {
		return false
	}

	if r.DirOnly && !isDir {
		if !directoryPatternMatches(r.Pattern, sub) {
			return false
		}
	}

	pattern := r.Pattern
	anchored := strings.HasPrefix(pattern, "/")
	pattern = strings.TrimPrefix(pattern, "/")

	if strings.Contains(pattern, "/") || anchored {
		fullPattern := pattern
		if r.Base != "" {
			fullPattern = path.Join(r.Base, pattern)
		}

		if ok, _ := doublestar.PathMatch(fullPattern, rel); ok {
			return true
		}

		if r.DirOnly && pathHasPrefix(rel, fullPattern) {
			return true
		}

		return false
	}

	parts := strings.Split(sub, "/")
	for i, part := range parts {
		if ok, _ := doublestar.PathMatch(pattern, part); ok {
			if r.DirOnly {
				prefix := strings.Join(parts[:i+1], "/")
				return pathHasPrefix(sub, prefix) || isDir
			}
			return true
		}
	}

	return false
}

func directoryPatternMatches(pattern string, sub string) bool {
	pattern = strings.TrimPrefix(pattern, "/")
	parts := strings.Split(sub, "/")

	for i, part := range parts {
		if ok, _ := doublestar.PathMatch(pattern, part); ok {
			return i < len(parts)-1
		}
	}

	if ok, _ := doublestar.PathMatch(pattern, sub); ok {
		return true
	}

	return pathHasPrefix(sub, pattern)
}

func cleanRel(rel string) string {
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." || rel == "" {
		return "."
	}
	return strings.TrimPrefix(rel, "./")
}

func pathHasPrefix(rel string, prefix string) bool {
	rel = cleanRel(rel)
	prefix = cleanRel(prefix)

	if prefix == "." {
		return true
	}

	return rel == prefix || strings.HasPrefix(rel, prefix+"/")
}

func pathDepth(p string) int {
	p = cleanRel(p)
	if p == "." || p == "" {
		return 0
	}
	return strings.Count(p, "/") + 1
}
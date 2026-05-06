package project

import (
	"context"

	"github.com/tedla-brandsema/mcpfs/internal/limits"
	fsservice "github.com/tedla-brandsema/mcpfs/internal/service/fs"
	gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"
)

const (
	defaultOverviewDepth         = 2
	maxOverviewDepth             = 5
	defaultOverviewEntries       = 500
	maxOverviewEntries           = 2000
	defaultOverviewRecentCommits = 5
	maxOverviewRecentCommits     = 20
)

func (s *Service) Overview(ctx context.Context, args OverviewArgs) (OverviewResult, error) {
	requestedPath := args.Path
	if requestedPath == "" {
		requestedPath = "."
	}

	maxDepth := limits.ClampInt(args.MaxDepth, defaultOverviewDepth, maxOverviewDepth)
	maxEntries := limits.ClampInt(args.MaxEntries, defaultOverviewEntries, maxOverviewEntries)
	recentCommits := limits.ClampInt(args.RecentCommits, defaultOverviewRecentCommits, maxOverviewRecentCommits)

	includeFiles := true
	tree, err := s.fs.Tree(ctx, fsservice.TreeArgs{
		RootID:       args.RootID,
		Path:         requestedPath,
		MaxDepth:     maxDepth,
		MaxEntries:   maxEntries,
		IncludeFiles: &includeFiles,
	})
	if err != nil {
		s.logDenied("project.overview", args.RootID, requestedPath, err.Error())
		return OverviewResult{}, err
	}

	result := OverviewResult{
		RootID:     tree.RootID,
		Path:       tree.Path,
		MaxDepth:   tree.MaxDepth,
		MaxEntries: tree.MaxEntries,
		TreeText:   tree.Text,
		TopLevel:   topLevelEntries(tree.Entries),
		Counts:     s.countEntries(tree.Entries),
		Truncated:  tree.Truncated,
	}

	result.ImportantFiles = s.importantFiles(tree.Entries)

	status, err := s.git.Status(ctx, gitservice.StatusArgs{
		RootID: args.RootID,
	})
	if err != nil {
		result.Git = OverviewGit{
			Available: false,
			Error:     err.Error(),
		}
		result.Warnings = append(result.Warnings, "git status unavailable: "+err.Error())
	} else {
		result.Git.Available = true
		result.Git.Branch = status.Branch
		result.Git.Clean = status.Clean
		result.Git.Changes = len(status.Changes)
		result.Git.Truncated = status.Truncated

		logResult, err := s.git.Log(ctx, gitservice.LogArgs{
			RootID:   args.RootID,
			Limit:    recentCommits,
			MaxBytes: 65536,
		})
		if err != nil {
			result.Warnings = append(result.Warnings, "git log unavailable: "+err.Error())
		} else {
			result.Git.RecentCommits = logResult.Commits
			result.Git.Truncated = result.Git.Truncated || logResult.Truncated
		}
	}

	s.logAllowed(
		"project.overview",
		result.RootID,
		result.Path,
		"entries", result.Counts.Entries,
		"important_files", len(result.ImportantFiles),
		"git_available", result.Git.Available,
		"truncated", result.Truncated,
	)

	return result, nil
}

func topLevelEntries(entries []fsservice.TreeEntry) []OverviewEntry {
	out := make([]OverviewEntry, 0)

	for _, entry := range entries {
		if entry.Depth != 1 {
			continue
		}

		out = append(out, OverviewEntry{
			Path:  entry.Path,
			Name:  entry.Name,
			Type:  entry.Type,
			Depth: entry.Depth,
			Size:  entry.Size,
		})
	}

	return out
}

func (s *Service) countEntries(entries []fsservice.TreeEntry) OverviewCounts {
	var counts OverviewCounts
	counts.Entries = len(entries)

	for _, entry := range entries {
		switch entry.Type {
		case "dir":
			counts.Directories++
		case "file":
			counts.Files++
			if s.registry.IsSourceFile(entry.Path) {
				counts.SourceFiles++
			}
			if s.registry.IsTestFile(entry.Path) {
				counts.TestFiles++
			}
			if s.registry.IsDocumentationFile(entry.Path) {
				counts.DocumentationFiles++
			}
			if s.registry.IsConfigurationFile(entry.Path) {
				counts.ConfigurationFiles++
			}
		}
	}

	return counts
}

func (s *Service) importantFiles(entries []fsservice.TreeEntry) []string {
	seen := make(map[string]bool)
	out := make([]string, 0)

	for _, entry := range entries {
		if entry.Type != "file" {
			continue
		}

		if !s.registry.IsImportantFile(entry.Path) {
			continue
		}

		if seen[entry.Path] {
			continue
		}

		seen[entry.Path] = true
		out = append(out, entry.Path)
	}

	return out
}

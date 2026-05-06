package project

import gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"

type OverviewArgs struct {
	RootID        string `json:"root_id" jsonschema:"configured root id"`
	Path          string `json:"path,omitempty" jsonschema:"relative directory path inside the root; defaults to ."`
	MaxDepth      int    `json:"max_depth,omitempty" jsonschema:"maximum tree depth to inspect"`
	MaxEntries    int    `json:"max_entries,omitempty" jsonschema:"maximum tree entries to inspect"`
	RecentCommits int    `json:"recent_commits,omitempty" jsonschema:"maximum recent commits to include"`
}

type OverviewResult struct {
	RootID         string          `json:"root_id"`
	Path           string          `json:"path"`
	MaxDepth       int             `json:"max_depth"`
	MaxEntries     int             `json:"max_entries"`
	TreeText       string          `json:"tree_text"`
	TopLevel       []OverviewEntry `json:"top_level"`
	ImportantFiles []string        `json:"important_files"`
	Counts         OverviewCounts  `json:"counts"`
	Git            OverviewGit     `json:"git"`
	Warnings       []string        `json:"warnings,omitempty"`
	Truncated      bool            `json:"truncated"`
}

type OverviewEntry struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Depth int    `json:"depth"`
	Size  int64  `json:"size,omitempty"`
}

type OverviewCounts struct {
	Entries            int `json:"entries"`
	Files              int `json:"files"`
	Directories        int `json:"directories"`
	SourceFiles        int `json:"source_files"`
	TestFiles          int `json:"test_files"`
	DocumentationFiles int `json:"documentation_files"`
	ConfigurationFiles int `json:"configuration_files"`
}

type OverviewGit struct {
	Available     bool                   `json:"available"`
	Branch        string                 `json:"branch,omitempty"`
	Clean         bool                   `json:"clean"`
	Changes       int                    `json:"changes"`
	RecentCommits []gitservice.LogCommit `json:"recent_commits,omitempty"`
	Truncated     bool                   `json:"truncated"`
	Error         string                 `json:"error,omitempty"`
}

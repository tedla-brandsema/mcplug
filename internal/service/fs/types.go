package fs

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
	RootID     string  `json:"root_id"`
	Path       string  `json:"path"`
	MaxEntries int     `json:"max_entries"`
	Entries    []Entry `json:"entries"`
	Truncated  bool    `json:"truncated"`
}

type TreeArgs struct {
	RootID       string `json:"root_id" jsonschema:"configured root id"`
	Path         string `json:"path,omitempty" jsonschema:"relative directory path inside the root"`
	MaxDepth     int    `json:"max_depth,omitempty" jsonschema:"maximum tree depth to return"`
	MaxEntries   int    `json:"max_entries,omitempty" jsonschema:"maximum number of tree entries to return"`
	IncludeFiles *bool  `json:"include_files,omitempty" jsonschema:"whether to include files in the tree output; defaults to true"`
}

type TreeEntry struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Depth      int    `json:"depth"`
	ParentPath string `json:"parent_path,omitempty"`
	Size       int64  `json:"size,omitempty"`
	MTime      string `json:"mtime,omitempty"`
}

type TreeResult struct {
	Root       TreeEntry   `json:"root"`
	RootID     string      `json:"root_id"`
	Path       string      `json:"path"`
	MaxDepth   int         `json:"max_depth"`
	MaxEntries int         `json:"max_entries"`
	Entries    []TreeEntry `json:"entries"`
	Text       string      `json:"text"`
	Truncated  bool        `json:"truncated"`
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
	Limit     int64  `json:"limit"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
}

type WriteArgs struct {
	RootID         string `json:"root_id" jsonschema:"configured root id"`
	Path           string `json:"path" jsonschema:"relative file path inside the root"`
	Content        string `json:"content" jsonschema:"file content to write"`
	CreateDirs     bool   `json:"create_dirs,omitempty" jsonschema:"whether to create parent directories"`
	ExpectedSHA256 string `json:"expected_sha256,omitempty" jsonschema:"optional SHA-256 hash the existing file must match before writing"`
}

type WriteResult struct {
	RootID string `json:"root_id"`
	Path   string `json:"path"`
	Bytes  int    `json:"bytes"`
	Mode   string `json:"mode"`
}

type PatchArgs struct {
	RootID         string      `json:"root_id" jsonschema:"configured root id"`
	Path           string      `json:"path" jsonschema:"relative file path inside the root"`
	Edits          []PatchEdit `json:"edits" jsonschema:"exact old/new text replacements to apply atomically"`
	DryRun         bool        `json:"dry_run,omitempty" jsonschema:"preview the patch without writing the file"`
	MaxDiffBytes   int         `json:"max_diff_bytes,omitempty" jsonschema:"maximum diff preview bytes to return"`
	ExpectedSHA256 string      `json:"expected_sha256,omitempty" jsonschema:"optional SHA-256 hash the existing file must match before patching"`
}

type PatchEdit struct {
	Old string `json:"old" jsonschema:"exact text block to replace; must match exactly once"`
	New string `json:"new" jsonschema:"replacement text block"`
}

type PatchResult struct {
	RootID        string `json:"root_id"`
	Path          string `json:"path"`
	Mode          string `json:"mode"`
	DryRun        bool   `json:"dry_run"`
	Changed       bool   `json:"changed"`
	EditsApplied  int    `json:"edits_applied"`
	BytesBefore   int    `json:"bytes_before"`
	BytesAfter    int    `json:"bytes_after"`
	MaxDiffBytes  int    `json:"max_diff_bytes"`
	Diff          string `json:"diff"`
	DiffTruncated bool   `json:"diff_truncated"`
}

type ReadLinesArgs struct {
	RootID    string `json:"root_id" jsonschema:"configured root id"`
	Path      string `json:"path" jsonschema:"relative file path inside the root"`
	StartLine int    `json:"start_line,omitempty" jsonschema:"1-based inclusive start line; defaults to 1"`
	EndLine   int    `json:"end_line,omitempty" jsonschema:"1-based inclusive end line; defaults to start_line plus the maximum line window"`
}

type ReadLine struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

type ReadLinesResult struct {
	RootID    string     `json:"root_id"`
	Path      string     `json:"path"`
	StartLine int        `json:"start_line"`
	EndLine   int        `json:"end_line"`
	MaxLines  int        `json:"max_lines"`
	Lines     []ReadLine `json:"lines"`
	Truncated bool       `json:"truncated"`
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
	RootID     string        `json:"root_id"`
	Query      string        `json:"query"`
	MaxResults int           `json:"max_results"`
	Matches    []SearchMatch `json:"matches"`
	Truncated  bool          `json:"truncated"`
}

type SearchRegexArgs struct {
	RootID        string `json:"root_id" jsonschema:"configured root id"`
	Query         string `json:"query" jsonschema:"regular expression query"`
	Glob          string `json:"glob,omitempty" jsonschema:"optional glob such as **/*.go"`
	CaseSensitive *bool  `json:"case_sensitive,omitempty" jsonschema:"whether regex matching is case-sensitive; defaults to true"`
	MaxResults    int    `json:"max_results,omitempty" jsonschema:"maximum number of matches"`
}

type SearchRegexResult struct {
	RootID        string        `json:"root_id"`
	Query         string        `json:"query"`
	Glob          string        `json:"glob,omitempty"`
	CaseSensitive bool          `json:"case_sensitive"`
	MaxResults    int           `json:"max_results"`
	Matches       []SearchMatch `json:"matches"`
	Truncated     bool          `json:"truncated"`
}

type HashArgs struct {
	RootID string `json:"root_id" jsonschema:"configured root id"`
	Path   string `json:"path" jsonschema:"relative file path inside the root"`
}

type HashResult struct {
	RootID string `json:"root_id"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	MTime  string `json:"mtime"`
	Mode   string `json:"mode"`
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

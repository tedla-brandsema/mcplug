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
package git

type StatusArgs struct {
	RootID string `json:"root_id" jsonschema:"configured root id"`
}

type StatusResult struct {
	RootID    string      `json:"root_id"`
	Branch    string      `json:"branch"`
	Clean     bool        `json:"clean"`
	Changes   []Change    `json:"changes"`
	Truncated bool        `json:"truncated"`
}

type Change struct {
	Path     string `json:"path"`
	OldPath  string `json:"old_path,omitempty"`
	Index    string `json:"index"`
	Worktree string `json:"worktree"`
	Staged   bool   `json:"staged"`
	Unstaged bool   `json:"unstaged"`
	Status   string `json:"status"`
}

type DiffArgs struct {
	RootID   string `json:"root_id" jsonschema:"configured root id"`
	Path     string `json:"path,omitempty" jsonschema:"optional relative path inside the root"`
	Staged   bool   `json:"staged,omitempty" jsonschema:"show staged diff instead of unstaged diff"`
	MaxBytes int    `json:"max_bytes,omitempty" jsonschema:"maximum diff bytes to return"`
}

type DiffResult struct {
	RootID    string `json:"root_id"`
	Path      string `json:"path,omitempty"`
	Staged    bool   `json:"staged"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
	Diff      string `json:"diff"`
}
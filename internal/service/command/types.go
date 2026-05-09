package command

type ListArgs struct{}

type CommandInfo struct {
	ID             string   `json:"id"`
	Description    string   `json:"description,omitempty"`
	RootID         string   `json:"root_id"`
	Workdir        string   `json:"workdir"`
	Command        []string `json:"command"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxOutputBytes int      `json:"max_output_bytes"`
}

type ListResult struct {
	Mode     string        `json:"mode"`
	Commands []CommandInfo `json:"commands"`
}

type RunArgs struct {
	ID             string `json:"id" jsonschema:"configured command id to run"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"optional timeout override in seconds"`
	MaxOutputBytes int    `json:"max_output_bytes,omitempty" jsonschema:"optional maximum combined stdout/stderr bytes to return"`
}

type RunResult struct {
	ID             string   `json:"id"`
	RootID         string   `json:"root_id"`
	Workdir        string   `json:"workdir"`
	Command        []string `json:"command"`
	ExitCode       int      `json:"exit_code"`
	DurationMS     int64    `json:"duration_ms"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxOutputBytes int      `json:"max_output_bytes"`
	Stdout         string   `json:"stdout"`
	Stderr         string   `json:"stderr"`
	Truncated      bool     `json:"truncated"`
	TimedOut       bool     `json:"timed_out"`
}

package git

import "testing"

func TestParseStatusCleanBranch(t *testing.T) {
	branch, changes, err := ParseStatus("## main...origin/main\n")
	if err != nil {
		t.Fatal(err)
	}

	if branch != "main...origin/main" {
		t.Fatalf("branch = %q, want %q", branch, "main...origin/main")
	}

	if len(changes) != 0 {
		t.Fatalf("changes len = %d, want 0", len(changes))
	}
}

func TestParseStatusModifiedUnstaged(t *testing.T) {
	_, changes, err := ParseStatus("## main\n M internal/service/fs/read.go\n")
	if err != nil {
		t.Fatal(err)
	}

	if len(changes) != 1 {
		t.Fatalf("changes len = %d, want 1", len(changes))
	}

	change := changes[0]
	if change.Path != "internal/service/fs/read.go" {
		t.Fatalf("path = %q", change.Path)
	}
	if change.Index != " " {
		t.Fatalf("index = %q, want space", change.Index)
	}
	if change.Worktree != "M" {
		t.Fatalf("worktree = %q, want M", change.Worktree)
	}
	if change.Staged {
		t.Fatal("expected staged=false")
	}
	if !change.Unstaged {
		t.Fatal("expected unstaged=true")
	}
	if change.Status != "modified" {
		t.Fatalf("status = %q, want modified", change.Status)
	}
}

func TestParseStatusAddedStaged(t *testing.T) {
	_, changes, err := ParseStatus("A  README.md\n")
	if err != nil {
		t.Fatal(err)
	}

	change := changes[0]
	if change.Path != "README.md" {
		t.Fatalf("path = %q", change.Path)
	}
	if !change.Staged {
		t.Fatal("expected staged=true")
	}
	if change.Unstaged {
		t.Fatal("expected unstaged=false")
	}
	if change.Status != "added" {
		t.Fatalf("status = %q, want added", change.Status)
	}
}

func TestParseStatusUntracked(t *testing.T) {
	_, changes, err := ParseStatus("?? internal/service/git/status.go\n")
	if err != nil {
		t.Fatal(err)
	}

	change := changes[0]
	if change.Path != "internal/service/git/status.go" {
		t.Fatalf("path = %q", change.Path)
	}
	if change.Staged {
		t.Fatal("expected staged=false")
	}
	if !change.Unstaged {
		t.Fatal("expected unstaged=true for untracked file")
	}
	if change.Status != "untracked" {
		t.Fatalf("status = %q, want untracked", change.Status)
	}
}

func TestParseStatusRenamed(t *testing.T) {
	_, changes, err := ParseStatus("R  old.go -> new.go\n")
	if err != nil {
		t.Fatal(err)
	}

	change := changes[0]
	if change.OldPath != "old.go" {
		t.Fatalf("old path = %q", change.OldPath)
	}
	if change.Path != "new.go" {
		t.Fatalf("path = %q", change.Path)
	}
	if change.Status != "renamed" {
		t.Fatalf("status = %q, want renamed", change.Status)
	}
}

func TestParseStatusMultiple(t *testing.T) {
	input := `## feature/git-tools
 M internal/service/fs/read.go
A  internal/service/git/status.go
?? internal/service/git/parse_test.go
`

	branch, changes, err := ParseStatus(input)
	if err != nil {
		t.Fatal(err)
	}

	if branch != "feature/git-tools" {
		t.Fatalf("branch = %q", branch)
	}

	if len(changes) != 3 {
		t.Fatalf("changes len = %d, want 3", len(changes))
	}
}

func TestParseStatusInvalidLine(t *testing.T) {
	_, _, err := ParseStatus("M\n")
	if err == nil {
		t.Fatal("expected error")
	}
}
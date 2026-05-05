package git

import (
	"bufio"
	"fmt"
	"strings"
)

func ParseStatus(output string) (string, []Change, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	branch := ""
	changes := make([]Change, 0)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			branch = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}

		change, err := parseStatusLine(line)
		if err != nil {
			return "", nil, err
		}

		changes = append(changes, change)
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	return branch, changes, nil
}

func parseStatusLine(line string) (Change, error) {
	if len(line) < 4 {
		return Change{}, fmt.Errorf("invalid git status line %q", line)
	}

	index := string(line[0])
	worktree := string(line[1])

	if line[2] != ' ' {
		return Change{}, fmt.Errorf("invalid git status separator in line %q", line)
	}

	pathText := strings.TrimSpace(line[3:])
	if pathText == "" {
		return Change{}, fmt.Errorf("missing path in git status line %q", line)
	}

	change := Change{
		Index:    index,
		Worktree: worktree,
		Staged:   isStaged(index),
		Unstaged: isUnstaged(worktree),
		Status:   describeStatus(index, worktree),
	}

	if strings.Contains(pathText, " -> ") {
		parts := strings.SplitN(pathText, " -> ", 2)
		change.OldPath = strings.TrimSpace(parts[0])
		change.Path = strings.TrimSpace(parts[1])
	} else {
		change.Path = pathText
	}

	return change, nil
}

func isStaged(index string) bool {
	return index != " " && index != "?" && index != "!"
}

func isUnstaged(worktree string) bool {
	return worktree != " " && worktree != "!"
}

func describeStatus(index string, worktree string) string {
	if index == "?" && worktree == "?" {
		return "untracked"
	}
	if index == "!" && worktree == "!" {
		return "ignored"
	}

	if index == "U" || worktree == "U" {
		return "unmerged"
	}

	if index == "R" || worktree == "R" {
		return "renamed"
	}
	if index == "C" || worktree == "C" {
		return "copied"
	}
	if index == "A" || worktree == "A" {
		return "added"
	}
	if index == "D" || worktree == "D" {
		return "deleted"
	}
	if index == "M" || worktree == "M" {
		return "modified"
	}

	return "changed"
}
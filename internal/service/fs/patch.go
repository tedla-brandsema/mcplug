package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

const defaultPatchDiffBytes = 65536

func (s *Service) Patch(ctx context.Context, args PatchArgs) (PatchResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.patch", args.RootID, args.Path, err.Error())
		return PatchResult{}, err
	}

	if root.Mode != config.ModeReadWrite {
		err := fmt.Errorf("root %q is not writable", root.ID)
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}

	if len(args.Edits) == 0 {
		err := fmt.Errorf("edits are required")
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}

	abs, err := core.ResolveWritableInsideRoot(root.RealPath, args.Path)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}

	rel, err := root.Rel(abs)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, args.Path, err.Error())
		return PatchResult{}, err
	}
	rel = cleanFSRel(rel)

	if !root.Matcher.AllowFile(rel) {
		err := fmt.Errorf("file is excluded")
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}
	if info.IsDir() {
		err := fmt.Errorf("path is a directory")
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}
	if info.Size() > root.MaxFileBytes {
		err := fmt.Errorf("file exceeds max_file_bytes: size=%d max=%d", info.Size(), root.MaxFileBytes)
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	if _, err := s.verifyExpectedSHA256(root, rel, args.ExpectedSHA256); err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	before := string(data)
	after, editsApplied, err := applyPatchEdits(before, args.Edits)
	if err != nil {
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	if int64(len(after)) > root.MaxFileBytes {
		err := fmt.Errorf("patched content exceeds max_file_bytes: size=%d max=%d", len(after), root.MaxFileBytes)
		s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
		return PatchResult{}, err
	}

	maxDiffBytes := args.MaxDiffBytes
	if maxDiffBytes <= 0 {
		maxDiffBytes = defaultPatchDiffBytes
	}

	diff := buildSimpleUnifiedDiff(rel, before, after)
	diffText, diffTruncated := limits.CapStringBytes(diff, maxDiffBytes)

	changed := before != after
	if changed && !args.DryRun {
		if err := os.WriteFile(abs, []byte(after), fs.FileMode(info.Mode().Perm())); err != nil {
			s.logDenied("mcpfs.patch", root.ID, rel, err.Error())
			return PatchResult{}, err
		}
	}

	result := PatchResult{
		RootID:        root.ID,
		Path:          rel,
		Mode:          string(root.Mode),
		DryRun:        args.DryRun,
		Changed:       changed,
		EditsApplied:  editsApplied,
		BytesBefore:   len(before),
		BytesAfter:    len(after),
		MaxDiffBytes:  maxDiffBytes,
		Diff:          diffText,
		DiffTruncated: diffTruncated,
	}

	s.logAllowed("mcpfs.patch", root.ID, rel, "edits_applied", result.EditsApplied, "changed", result.Changed, "dry_run", result.DryRun)
	return result, nil
}

func applyPatchEdits(content string, edits []PatchEdit) (string, int, error) {
	patched := content

	for i, edit := range edits {
		if edit.Old == "" {
			return "", 0, fmt.Errorf("edits[%d].old must not be empty", i)
		}

		count := strings.Count(patched, edit.Old)
		if count == 0 {
			return "", 0, fmt.Errorf("edits[%d].old matched 0 times", i)
		}
		if count > 1 {
			return "", 0, fmt.Errorf("edits[%d].old matched %d times", i, count)
		}

		patched = strings.Replace(patched, edit.Old, edit.New, 1)
	}

	return patched, len(edits), nil
}

func buildSimpleUnifiedDiff(path string, before string, after string) string {
	if before == after {
		return ""
	}

	var b strings.Builder
	b.WriteString("--- a/")
	b.WriteString(path)
	b.WriteString("\n")
	b.WriteString("+++ b/")
	b.WriteString(path)
	b.WriteString("\n")

	beforeLines := splitPatchLines(before)
	afterLines := splitPatchLines(after)
	b.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(beforeLines), len(afterLines)))

	for _, line := range beforeLines {
		b.WriteString("-")
		b.WriteString(line)
	}
	for _, line := range afterLines {
		b.WriteString("+")
		b.WriteString(line)
	}

	return b.String()
}

func splitPatchLines(content string) []string {
	if content == "" {
		return nil
	}

	lines := strings.SplitAfter(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

package fs

import (
	"context"
	"fmt"
	iofs "io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/tedla-brandsema/mcpfs/internal/core"
	"github.com/tedla-brandsema/mcpfs/internal/limits"
)

type treeBuildNode struct {
	Entry    TreeEntry
	Children []treeBuildNode
}

func (s *Service) Tree(ctx context.Context, args TreeArgs) (TreeResult, error) {
	_ = ctx

	root, err := s.root(args.RootID)
	if err != nil {
		s.logDenied("mcpfs.tree", args.RootID, args.Path, err.Error())
		return TreeResult{}, err
	}

	requested := args.Path
	if requested == "" {
		requested = "."
	}

	rel, err := s.resolve(root, requested)
	if err != nil {
		s.logDenied("mcpfs.tree", root.ID, requested, err.Error())
		return TreeResult{}, err
	}

	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		s.logDenied("mcpfs.tree", root.ID, rel, err.Error())
		return TreeResult{}, err
	}
	if !info.IsDir() {
		err := fmt.Errorf("path is not a directory")
		s.logDenied("mcpfs.tree", root.ID, rel, err.Error())
		return TreeResult{}, err
	}
	if !root.Matcher.AllowDir(rel) {
		err := fmt.Errorf("directory is excluded")
		s.logDenied("mcpfs.tree", root.ID, rel, err.Error())
		return TreeResult{}, err
	}

	maxDepth := limits.ClampInt(args.MaxDepth, 3, 10)
	maxEntries := limits.ClampInt(args.MaxEntries, 300, 1000)

	includeFiles := true
	if args.IncludeFiles != nil {
		includeFiles = *args.IncludeFiles
	}

	rootEntry, err := makeTreeEntry(root, rel, 0, "")
	if err != nil {
		s.logDenied("mcpfs.tree", root.ID, rel, err.Error())
		return TreeResult{}, err
	}

	tree := treeBuildNode{
		Entry: rootEntry,
	}

	entries := make([]TreeEntry, 0)
	count := 0
	truncated := false

	err = s.populateTree(root, &tree, rel, 0, maxDepth, maxEntries, includeFiles, &entries, &count, &truncated)
	if err != nil {
		s.logDenied("mcpfs.tree", root.ID, rel, err.Error())
		return TreeResult{}, err
	}

	result := TreeResult{
		RootID:    root.ID,
		Path:      rel,
		Root:      rootEntry,
		Entries:   entries,
		Text:      renderTreeText(tree),
		Truncated: truncated,
	}

	s.logAllowed("mcpfs.tree", root.ID, rel, "entries", count, "truncated", result.Truncated)
	return result, nil
}

func (s *Service) populateTree(
	root *core.Root,
	node *treeBuildNode,
	rel string,
	depth int,
	maxDepth int,
	maxEntries int,
	includeFiles bool,
	entries *[]TreeEntry,
	count *int,
	truncated *bool,
) error {
	if depth >= maxDepth {
		return nil
	}

	dirEntries, err := iofs.ReadDir(root.ReadFS, rel)
	if err != nil {
		return err
	}

	for _, d := range dirEntries {
		if *count >= maxEntries {
			*truncated = true
			break
		}

		entryRel := joinRel(rel, d.Name())

		safeRel, err := s.resolve(root, entryRel)
		if err != nil {
			s.logDenied("mcpfs.tree", root.ID, entryRel, err.Error())
			continue
		}

		if d.IsDir() {
			if !root.Matcher.AllowDir(safeRel) {
				continue
			}
		} else {
			if !includeFiles || !root.Matcher.AllowFile(safeRel) {
				continue
			}
		}

		childEntry, err := makeTreeEntry(root, safeRel, depth+1, rel)
		if err != nil {
			continue
		}

		child := treeBuildNode{
			Entry: childEntry,
		}

		node.Children = append(node.Children, child)
		*entries = append(*entries, childEntry)
		(*count)++

		if d.IsDir() {
			idx := len(node.Children) - 1
			if err := s.populateTree(root, &node.Children[idx], safeRel, depth+1, maxDepth, maxEntries, includeFiles, entries, count, truncated); err != nil {
				return err
			}
		}
	}

	return nil
}

func makeTreeEntry(root *core.Root, rel string, depth int, parentPath string) (TreeEntry, error) {
	info, err := iofs.Stat(root.ReadFS, rel)
	if err != nil {
		return TreeEntry{}, err
	}

	typ := "file"
	if info.IsDir() {
		typ = "dir"
	}

	name := path.Base(filepath.ToSlash(rel))
	if rel == "." {
		name = "."
	}

	return TreeEntry{
		Path:       filepath.ToSlash(rel),
		Name:       name,
		Type:       typ,
		Depth:      depth,
		ParentPath: parentPath,
		Size:       info.Size(),
		MTime:      info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

func renderTreeText(root treeBuildNode) string {
	var b strings.Builder
	b.WriteString(treeTextName(root.Entry))
	b.WriteByte('\n')
	renderTreeChildren(&b, root.Children, "")
	return strings.TrimRight(b.String(), "\n")
}

func renderTreeChildren(b *strings.Builder, children []treeBuildNode, prefix string) {
	for i, child := range children {
		last := i == len(children)-1

		connector := "├── "
		nextPrefix := prefix + "│   "
		if last {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		b.WriteString(prefix)
		b.WriteString(connector)
		b.WriteString(treeTextName(child.Entry))
		b.WriteByte('\n')

		if len(child.Children) > 0 {
			renderTreeChildren(b, child.Children, nextPrefix)
		}
	}
}

func treeTextName(entry TreeEntry) string {
	if entry.Path == "." {
		return "."
	}
	if entry.Name != "" {
		return entry.Name
	}
	return entry.Path
}

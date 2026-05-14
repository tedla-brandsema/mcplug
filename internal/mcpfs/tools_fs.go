package mcpfs

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	fsservice "github.com/tedla-brandsema/mcpfs/internal/service/fs"
)

func RegisterFSTools(server *mcp.Server, svc *fsservice.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_roots",
		Description: "List configured filesystem roots and their read modes. Does not expose absolute host paths.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.RootsArgs) (*mcp.CallToolResult, fsservice.RootsResult, error) {
		result, err := svc.Roots(ctx, args)
		if err != nil {
			return toolError(err), fsservice.RootsResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_list",
		Description: "List files under a configured filesystem root. Path must be relative to the selected root. Honors explicit excludes and .gitignore rules.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.ListArgs) (*mcp.CallToolResult, fsservice.ListResult, error) {
		result, err := svc.List(ctx, args)
		if err != nil {
			return toolError(err), fsservice.ListResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_tree",
		Description: "Return a bounded tree view under a configured filesystem root. Honors explicit excludes, .gitignore rules, and symlink checks.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.TreeArgs) (*mcp.CallToolResult, fsservice.TreeResult, error) {
		result, err := svc.Tree(ctx, args)
		if err != nil {
			return toolError(err), fsservice.TreeResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_read",
		Description: "Read a file from a configured filesystem root. Path must be relative to the selected root. Honors explicit excludes, .gitignore rules, symlink checks, and file size limits.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.ReadArgs) (*mcp.CallToolResult, fsservice.ReadResult, error) {
		result, err := svc.Read(ctx, args)
		if err != nil {
			return toolError(err), fsservice.ReadResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_read_lines",
		Description: "Read a 1-based inclusive line range from a file under a configured root. Honors explicit excludes, .gitignore rules, symlink checks, and file size limits.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.ReadLinesArgs) (*mcp.CallToolResult, fsservice.ReadLinesResult, error) {
		result, err := svc.ReadLines(ctx, args)
		if err != nil {
			return toolError(err), fsservice.ReadLinesResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_write",
		Description: "Create or replace a file under a configured read_write filesystem root. Path must be relative to the selected root. Honors explicit excludes, .gitignore rules, symlink checks, and file size limits.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: false,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.WriteArgs) (*mcp.CallToolResult, fsservice.WriteResult, error) {
		result, err := svc.Write(ctx, args)
		if err != nil {
			return toolError(err), fsservice.WriteResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_patch",
		Description: "Apply exact old/new text replacements to an existing file under a configured read_write filesystem root. Each old block must match exactly once. Supports dry_run diff previews.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: false,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.PatchArgs) (*mcp.CallToolResult, fsservice.PatchResult, error) {
		result, err := svc.Patch(ctx, args)
		if err != nil {
			return toolError(err), fsservice.PatchResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_search",
		Description: "Search text files under a configured filesystem root using a case-sensitive substring query. Honors explicit excludes and .gitignore rules.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.SearchArgs) (*mcp.CallToolResult, fsservice.SearchResult, error) {
		result, err := svc.Search(ctx, args)
		if err != nil {
			return toolError(err), fsservice.SearchResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_search_regex",
		Description: "Search text files under a configured filesystem root using a regular expression query. Honors explicit excludes and .gitignore rules.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.SearchRegexArgs) (*mcp.CallToolResult, fsservice.SearchRegexResult, error) {
		result, err := svc.SearchRegex(ctx, args)
		if err != nil {
			return toolError(err), fsservice.SearchRegexResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_stat",
		Description: "Return metadata for a file or directory under a configured filesystem root. Path must be relative to the selected root.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.StatArgs) (*mcp.CallToolResult, fsservice.StatResult, error) {
		result, err := svc.Stat(ctx, args)
		if err != nil {
			return toolError(err), fsservice.StatResult{}, nil
		}
		return toolJSON(result), result, nil
	})
}

func toolJSON(v any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: jsonString(v)},
		},
	}
}

func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}
}

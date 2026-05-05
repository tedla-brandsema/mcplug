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
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.ListArgs) (*mcp.CallToolResult, fsservice.ListResult, error) {
		result, err := svc.List(ctx, args)
		if err != nil {
			return toolError(err), fsservice.ListResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_read",
		Description: "Read a file from a configured filesystem root. Path must be relative to the selected root. Honors explicit excludes, .gitignore rules, symlink checks, and file size limits.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.ReadArgs) (*mcp.CallToolResult, fsservice.ReadResult, error) {
		result, err := svc.Read(ctx, args)
		if err != nil {
			return toolError(err), fsservice.ReadResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_search",
		Description: "Search text files under a configured filesystem root using a case-sensitive substring query. Honors explicit excludes and .gitignore rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fsservice.SearchArgs) (*mcp.CallToolResult, fsservice.SearchResult, error) {
		result, err := svc.Search(ctx, args)
		if err != nil {
			return toolError(err), fsservice.SearchResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_stat",
		Description: "Return metadata for a file or directory under a configured filesystem root. Path must be relative to the selected root.",
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
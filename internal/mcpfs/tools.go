package mcpfs

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterTools(server *mcp.Server, svc *Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_roots",
		Description: "List configured filesystem roots and their read modes. Does not expose absolute host paths.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RootsArgs) (*mcp.CallToolResult, RootsResult, error) {
		result, err := svc.Roots(ctx, args)
		if err != nil {
			return toolError(err), RootsResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_list",
		Description: "List files under a configured filesystem root. Path must be relative to the selected root. Honors explicit excludes and .gitignore rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListArgs) (*mcp.CallToolResult, ListResult, error) {
		result, err := svc.List(ctx, args)
		if err != nil {
			return toolError(err), ListResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_read",
		Description: "Read a file from a configured filesystem root. Path must be relative to the selected root. Honors explicit excludes, .gitignore rules, symlink checks, and file size limits.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ReadArgs) (*mcp.CallToolResult, ReadResult, error) {
		result, err := svc.Read(ctx, args)
		if err != nil {
			return toolError(err), ReadResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_search",
		Description: "Search text files under a configured filesystem root using a case-sensitive substring query. Honors explicit excludes and .gitignore rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, SearchResult, error) {
		result, err := svc.Search(ctx, args)
		if err != nil {
			return toolError(err), SearchResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_stat",
		Description: "Return metadata for a file or directory under a configured filesystem root. Path must be relative to the selected root.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args StatArgs) (*mcp.CallToolResult, StatResult, error) {
		result, err := svc.Stat(ctx, args)
		if err != nil {
			return toolError(err), StatResult{}, nil
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
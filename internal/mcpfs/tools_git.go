package mcpfs

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"
)

func RegisterGitTools(server *mcp.Server, svc *gitservice.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "git_status",
		Description: "Return read-only git status for a configured filesystem root using git status --porcelain=v1 -b.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args gitservice.StatusArgs) (*mcp.CallToolResult, gitservice.StatusResult, error) {
		result, err := svc.Status(ctx, args)
		if err != nil {
			return toolError(err), gitservice.StatusResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "git_diff",
		Description: "Return read-only git diff for a configured filesystem root. Optionally restrict to a relative path. Uses git diff or git diff --cached.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args gitservice.DiffArgs) (*mcp.CallToolResult, gitservice.DiffResult, error) {
		result, err := svc.Diff(ctx, args)
		if err != nil {
			return toolError(err), gitservice.DiffResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "git_show",
		Description: "Return read-only metadata and patch for a single git commit. Optionally restrict the patch to a relative path.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args gitservice.ShowArgs) (*mcp.CallToolResult, gitservice.ShowResult, error) {
		result, err := svc.Show(ctx, args)
		if err != nil {
			return toolError(err), gitservice.ShowResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "git_log",
		Description: "Return recent git commit history for a configured filesystem root. Optionally restrict to a path.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args gitservice.LogArgs) (*mcp.CallToolResult, gitservice.LogResult, error) {
		result, err := svc.Log(ctx, args)
		if err != nil {
			return toolError(err), gitservice.LogResult{}, nil
		}
		return toolJSON(result), result, nil
	})
}

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
	}, func(ctx context.Context, req *mcp.CallToolRequest, args gitservice.DiffArgs) (*mcp.CallToolResult, gitservice.DiffResult, error) {
		result, err := svc.Diff(ctx, args)
		if err != nil {
			return toolError(err), gitservice.DiffResult{}, nil
		}
		return toolJSON(result), result, nil
	})
}
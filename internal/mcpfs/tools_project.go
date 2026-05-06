package mcpfs

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	projectservice "github.com/tedla-brandsema/mcpfs/internal/service/project"
)

func RegisterProjectTools(server *mcp.Server, svc *projectservice.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_overview",
		Description: "Return a compact, bounded overview of a configured project root, including tree summary, important files, counts, git status, and recent commits.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args projectservice.OverviewArgs) (*mcp.CallToolResult, projectservice.OverviewResult, error) {
		result, err := svc.Overview(ctx, args)
		if err != nil {
			return toolError(err), projectservice.OverviewResult{}, nil
		}
		return toolJSON(result), result, nil
	})
}

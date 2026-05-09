package mcpfs

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	commandservice "github.com/tedla-brandsema/mcpfs/internal/service/command"
)

func RegisterCommandTools(server *mcp.Server, svc *commandservice.Service) {
	if !svc.Enabled() {
		return
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cmd_list",
		Description: "List configured command IDs available through MCPFS command execution.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args commandservice.ListArgs) (*mcp.CallToolResult, commandservice.ListResult, error) {
		result, err := svc.List(ctx, args)
		if err != nil {
			return toolError(err), commandservice.ListResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cmd_run",
		Description: "Run a predefined command by configured command ID. Commands execute with fixed argv, root-scoped workdir, timeout, and output limits.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: false,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args commandservice.RunArgs) (*mcp.CallToolResult, commandservice.RunResult, error) {
		result, err := svc.Run(ctx, args)
		if err != nil {
			return toolError(err), commandservice.RunResult{}, nil
		}
		return toolJSON(result), result, nil
	})

	if !svc.Unguarded() {
		return
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cmd_exec",
		Description: "Run an arbitrary argv command in unguarded command mode. Treat this like terminal access. Commands execute in a root-scoped workdir with timeout and output limits.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: false,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args commandservice.ExecArgs) (*mcp.CallToolResult, commandservice.ExecResult, error) {
		result, err := svc.Exec(ctx, args)
		if err != nil {
			return toolError(err), commandservice.ExecResult{}, nil
		}
		return toolJSON(result), result, nil
	})
}

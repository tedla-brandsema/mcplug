package gateway

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the aggregated MCP server. Construct it with BuildServer.
type Server struct {
	MCP *mcp.Server
}

package mcpfs

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	gitservice "github.com/tedla-brandsema/mcpfs/internal/service/git"
)

func RegisterGitTools(server *mcp.Server, svc *gitservice.Service) {
	_ = server
	_ = svc
}
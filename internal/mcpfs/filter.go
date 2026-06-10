package mcpfs

import "github.com/modelcontextprotocol/go-sdk/mcp"

// FilterTools applies the per-server allow/deny lists. With include set, only
// listed tools pass and missing reports include entries the upstream did not
// offer; otherwise tools in exclude are dropped. Config validation guarantees
// include and exclude are never both set.
func FilterTools(tools []*mcp.Tool, include, exclude []string) (filtered []*mcp.Tool, missing []string) {
	if len(include) > 0 {
		byName := make(map[string]*mcp.Tool, len(tools))
		for _, t := range tools {
			byName[t.Name] = t
		}
		for _, name := range include {
			if t, ok := byName[name]; ok {
				filtered = append(filtered, t)
			} else {
				missing = append(missing, name)
			}
		}
		return filtered, missing
	}

	if len(exclude) == 0 {
		return tools, nil
	}

	excluded := make(map[string]struct{}, len(exclude))
	for _, name := range exclude {
		excluded[name] = struct{}{}
	}
	for _, t := range tools {
		if _, ok := excluded[t.Name]; !ok {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

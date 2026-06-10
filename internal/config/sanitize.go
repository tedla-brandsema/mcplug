package config

import (
	"fmt"
	"strings"
)

// SanitizeServerName converts a configured mcpServers name into the prefix
// used for exposed tool names (<prefix>_<tool>):
//
//   - characters outside [A-Za-z0-9_-] become "_"
//   - runs of "_" collapse to one
//   - leading/trailing "_" are trimmed
//   - an empty result is an error
//   - a result not starting with a letter is prefixed with "server_"
func SanitizeServerName(name string) (string, error) {
	var b strings.Builder
	b.Grow(len(name))

	prevUnderscore := false
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
			prevUnderscore = false
		default:
			if !prevUnderscore {
				b.WriteByte('_')
				prevUnderscore = true
			}
		}
	}

	sanitized := strings.Trim(b.String(), "_")
	if sanitized == "" {
		return "", fmt.Errorf("name %q sanitizes to an empty tool prefix", name)
	}

	first := sanitized[0]
	if !(first >= 'A' && first <= 'Z' || first >= 'a' && first <= 'z') {
		sanitized = "server_" + sanitized
	}

	return sanitized, nil
}

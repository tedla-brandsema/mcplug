package mcpfs

import "log/slog"

func (s *Service) logAllowed(event string, rootID string, path string, attrs ...any) {
	args := []any{
		"event", event,
		"root_id", rootID,
		"path", path,
	}
	args = append(args, attrs...)
	s.logger.Info("mcpfs allowed", args...)
}

func (s *Service) logDenied(event string, rootID string, path string, reason string) {
	s.logger.Warn(
		"mcpfs denied",
		slog.String("event", event),
		slog.String("root_id", rootID),
		slog.String("path", path),
		slog.String("reason", reason),
	)
}
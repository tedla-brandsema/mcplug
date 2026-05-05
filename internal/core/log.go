package core

import "log/slog"

func LogAllowed(logger *slog.Logger, service Service, event string, rootID string, path string, attrs ...any) {
	if logger == nil {
		logger = slog.Default()
	}

	args := []any{
		"service", serviceName(service),
		"event", event,
		"root_id", rootID,
		"path", path,
	}
	args = append(args, attrs...)

	logger.Info("mcpfs allowed", args...)
}

func LogDenied(logger *slog.Logger, service Service, event string, rootID string, path string, reason string) {
	if logger == nil {
		logger = slog.Default()
	}

	logger.Warn(
		"mcpfs denied",
		slog.String("service", serviceName(service)),
		slog.String("event", event),
		slog.String("root_id", rootID),
		slog.String("path", path),
		slog.String("reason", reason),
	)
}

func serviceName(service Service) string {
	if service == nil {
		return "unknown"
	}
	return service.Name()
}
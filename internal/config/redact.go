package config

import "strings"

var secretKeyMarkers = []string{
	"token",
	"authorization",
	"api_key",
	"apikey",
	"secret",
	"password",
	"bearer",
}

const RedactedValue = "[redacted]"

// IsSecretKey reports whether a header or env key likely names a secret.
// Dashes are treated as underscores so header-style keys like X-Api-Key match.
func IsSecretKey(key string) bool {
	normalized := strings.ReplaceAll(strings.ToLower(key), "-", "_")
	for _, marker := range secretKeyMarkers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

// RedactValue returns the value unchanged unless its key likely names a
// secret, in which case it returns RedactedValue. Header and env values must
// never be logged directly; use this for any config echo.
func RedactValue(key, value string) string {
	if IsSecretKey(key) {
		return RedactedValue
	}
	return value
}

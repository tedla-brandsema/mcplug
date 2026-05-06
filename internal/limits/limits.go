package limits

func ClampInt(requested int, defaultValue int, maxValue int) int {
	if requested <= 0 {
		return defaultValue
	}
	if maxValue > 0 && requested > maxValue {
		return maxValue
	}
	return requested
}

func ClampInt64(requested int64, defaultValue int64, maxValue int64) int64 {
	if requested <= 0 {
		return defaultValue
	}
	if maxValue > 0 && requested > maxValue {
		return maxValue
	}
	return requested
}

func CapStringBytes(s string, maxBytes int) (string, bool) {
	if maxBytes <= 0 {
		return "", len(s) > 0
	}
	if len(s) <= maxBytes {
		return s, false
	}
	return s[:maxBytes], true
}
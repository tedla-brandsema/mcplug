package limits

import "testing"

func TestClampIntUsesDefaultForZeroOrNegative(t *testing.T) {
	if got := ClampInt(0, 10, 100); got != 10 {
		t.Fatalf("ClampInt(0, 10, 100) = %d, want 10", got)
	}
	if got := ClampInt(-1, 10, 100); got != 10 {
		t.Fatalf("ClampInt(-1, 10, 100) = %d, want 10", got)
	}
}

func TestClampIntCapsAtMax(t *testing.T) {
	if got := ClampInt(200, 10, 100); got != 100 {
		t.Fatalf("ClampInt(200, 10, 100) = %d, want 100", got)
	}
}

func TestClampIntAllowsValueWithinRange(t *testing.T) {
	if got := ClampInt(50, 10, 100); got != 50 {
		t.Fatalf("ClampInt(50, 10, 100) = %d, want 50", got)
	}
}

func TestClampIntAllowsUnlimitedMaxWhenMaxIsZero(t *testing.T) {
	if got := ClampInt(200, 10, 0); got != 200 {
		t.Fatalf("ClampInt(200, 10, 0) = %d, want 200", got)
	}
}

func TestClampInt64UsesDefaultForZeroOrNegative(t *testing.T) {
	if got := ClampInt64(0, 10, 100); got != 10 {
		t.Fatalf("ClampInt64(0, 10, 100) = %d, want 10", got)
	}
	if got := ClampInt64(-1, 10, 100); got != 10 {
		t.Fatalf("ClampInt64(-1, 10, 100) = %d, want 10", got)
	}
}

func TestClampInt64CapsAtMax(t *testing.T) {
	if got := ClampInt64(200, 10, 100); got != 100 {
		t.Fatalf("ClampInt64(200, 10, 100) = %d, want 100", got)
	}
}

func TestCapStringBytesTruncates(t *testing.T) {
	got, truncated := CapStringBytes("hello world", 5)
	if got != "hello" {
		t.Fatalf("got = %q, want %q", got, "hello")
	}
	if !truncated {
		t.Fatal("truncated = false, want true")
	}
}

func TestCapStringBytesNoTruncation(t *testing.T) {
	got, truncated := CapStringBytes("hello", 10)
	if got != "hello" {
		t.Fatalf("got = %q, want %q", got, "hello")
	}
	if truncated {
		t.Fatal("truncated = true, want false")
	}
}

func TestCapStringBytesZeroLimit(t *testing.T) {
	got, truncated := CapStringBytes("hello", 0)
	if got != "" {
		t.Fatalf("got = %q, want empty string", got)
	}
	if !truncated {
		t.Fatal("truncated = false, want true")
	}
}

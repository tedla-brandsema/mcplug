package git

import (
	"strings"
	"testing"
)

func TestCappedBufferHonorsLimit(t *testing.T) {
	var b cappedBuffer
	b.limit = 5

	n, err := b.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len("hello world") {
		t.Fatalf("n = %d, want %d", n, len("hello world"))
	}
	if got := b.String(); got != "hello" {
		t.Fatalf("String = %q, want %q", got, "hello")
	}
	if !b.truncated {
		t.Fatal("truncated = false, want true")
	}
}

func TestCappedBufferMultipleWrites(t *testing.T) {
	var b cappedBuffer
	b.limit = 8

	_, _ = b.Write([]byte("hello"))
	_, _ = b.Write([]byte(" world"))

	if got := b.String(); got != "hello wo" {
		t.Fatalf("String = %q, want %q", got, "hello wo")
	}
	if !b.truncated {
		t.Fatal("truncated = false, want true")
	}
}

func TestCappedBufferNoTruncation(t *testing.T) {
	var b cappedBuffer
	b.limit = 20

	_, _ = b.Write([]byte(strings.Repeat("a", 10)))

	if got := b.String(); len(got) != 10 {
		t.Fatalf("len(String) = %d, want 10", len(got))
	}
	if b.truncated {
		t.Fatal("truncated = true, want false")
	}
}
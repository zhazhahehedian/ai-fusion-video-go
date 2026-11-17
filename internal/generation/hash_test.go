package generation

import "testing"

func TestScopeKeyStable(t *testing.T) {
	a := ScopeKey(42)
	b := ScopeKey(42)
	c := ScopeKey(43)
	if a == "" || a != b {
		t.Fatalf("scope key unstable: %q vs %q", a, b)
	}
	if a == c {
		t.Fatalf("different users must not share scope key")
	}
}

func TestRequestHashConflictDetection(t *testing.T) {
	h1 := RequestHash(KindEcho, "hello", 1)
	h2 := RequestHash(KindEcho, "hello", 1)
	h3 := RequestHash(KindEcho, "world", 1)
	if h1 != h2 {
		t.Fatalf("same request must hash equal")
	}
	if h1 == h3 {
		t.Fatalf("different prompt must hash different")
	}
}

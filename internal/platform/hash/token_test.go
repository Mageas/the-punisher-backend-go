package hash

import "testing"

func TestHashToken(t *testing.T) {
	token := "refresh-token"
	secret := "secret"

	h1 := HashToken(token, secret)
	h2 := HashToken(token, secret)
	h3 := HashToken(token, "another-secret")

	if h1 != h2 {
		t.Fatalf("hash should be deterministic")
	}
	if h1 == h3 {
		t.Fatalf("hash should change when secret changes")
	}
	if len(h1) != 64 {
		t.Fatalf("expected sha256 hex length 64, got %d", len(h1))
	}
}

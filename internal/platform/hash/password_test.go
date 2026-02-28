package hash

import "testing"

func TestHashPasswordAndVerify(t *testing.T) {
	plain := "Sup3rS3cret!"

	hashed, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hashed == plain {
		t.Fatalf("hashed password should differ from plain text")
	}

	if err := VerifyPassword(plain, hashed); err != nil {
		t.Fatalf("VerifyPassword should accept the original password: %v", err)
	}

	if err := VerifyPassword("wrong-password", hashed); err == nil {
		t.Fatalf("VerifyPassword should reject an invalid password")
	}
}

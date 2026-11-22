package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "this_is_my_password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned an error : %v\n", err)
	}

	if hash == "" {
		t.Fatal("HashPassword returned an empty hash")
	}

	if hash == password {
		t.Fatal("hash must not equal to plaintext password")
	}
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	password := "this_is_my_password"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Fatal("expected different hashes for the same password (random salt)")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
		wantErr  bool
	}{
		{"correct password", password, hash, true, false},
		{"wrong password", "wrong-password", hash, false, false},
		{"empty password", "", hash, false, false},
		{"malformed hash", password, "not-a-valid-hash", false, true},
		{"empty hash", password, "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

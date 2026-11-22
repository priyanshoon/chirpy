package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "super-secret-key"
	duration := time.Hour

	tokenStr, err := MakeJWT(userID, secret, duration)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}

	parts := len(tokenStr)
	if parts < 10 {
		t.Fatalf("token string looks too short to be a vaild jwt: %s", tokenStr)
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "super-secret-key"
	duration := time.Hour

	// 1. Happy Path: Valid token and correct secret
	t.Run("Valid token", func(t *testing.T) {
		tokenStr, _ := MakeJWT(userID, secret, duration)

		parsedID, err := ValidateJWT(tokenStr, secret)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if parsedID != userID {
			t.Fatalf("expected userID %v, got %v", userID, parsedID)
		}
	})

	// 2. Failure Path: Valid token, but WRONG secret
	t.Run("Invalid token", func(t *testing.T) {
		tokenStr, _ := MakeJWT(userID, secret, duration)
		_, err := ValidateJWT(tokenStr, "wrong-secret")
		if err == nil {
			t.Fatal("expected an error for wrong-secret, got nil")
		}
	})

	// 3. Failure Path: Expired token
	t.Run("Expired Token", func(t *testing.T) {
		// Create a token that expired 1 hour ago (negative duration)
		tokenStr, _ := MakeJWT(userID, secret, -time.Hour)

		_, err := ValidateJWT(tokenStr, secret)
		if err == nil {
			t.Fatal("expected an error for expired token, got nil")
		}
	})

	// 4. Failure Path: Malformed/Garbage token string
	t.Run("Malformed Token", func(t *testing.T) {
		_, err := ValidateJWT("this.is.not.a.real.jwt", secret)
		if err == nil {
			t.Fatal("expected an error for malformed token, got nil")
		}
	})

	// 5. Failure Path: Empty token string
	t.Run("Empty Token", func(t *testing.T) {
		_, err := ValidateJWT("", secret)
		if err == nil {
			t.Fatal("expected an error for empty token, got nil")
		}
	})
}

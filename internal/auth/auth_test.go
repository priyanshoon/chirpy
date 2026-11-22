package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetBearerToken(t *testing.T) {
	userID := uuid.New()
	header := make(http.Header)
	token, _ := MakeJWT(userID, "token_secret", time.Hour)
	header.Set("Authorization", "Bearer "+token)

	fetchedToken, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	if fetchedToken == "" {
		t.Fatal("expected non empty string")
	}

	if fetchedToken != token {
		t.Fatalf("expected %v, got %v", token, fetchedToken)
	}

	if strings.HasPrefix(fetchedToken, "Bearer ") {
		t.Fatal("expected not to have Bearer Prefix")
	}
}

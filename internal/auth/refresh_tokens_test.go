package auth

import (
	"testing"
)

func TestMakeRefreshToken(t *testing.T) {
	got := MakeRefreshToken()

	if got == "" {
		t.Fatal("expected non empty string")
	}

	if len(got) < 64 {
		t.Fatal("expected 64 char string")
	}
}

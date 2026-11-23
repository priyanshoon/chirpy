package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() string {
	b := make([]byte, 32)

	rand.Read(b)

	hexS := hex.EncodeToString(b)
	return hexS
}

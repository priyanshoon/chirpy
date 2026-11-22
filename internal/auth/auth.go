package auth

import (
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return "", http.ErrMissingFile
	}

	return strings.TrimPrefix(authorizationHeader, "Bearer "), nil
}

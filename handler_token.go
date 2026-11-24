package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/priyanshoon/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefersh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "cannot get jwt", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "something went wrong", err)
		return
	}

	newAccessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "cannot make jwt token, something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: newAccessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "auth header is missing", errors.New("auth header is missing"))
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusBadRequest, "auth header is invalid", errors.New("auth header is invalid"))
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	err := cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

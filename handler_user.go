package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshoon/chirpy/internal/auth"
	"github.com/priyanshoon/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token,omitempty"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "cannot create a hash password", err)
	}

	sendUserParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), sendUserParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password        string `json:"password"`
		Email           string `json:"email"`
		ExpiresInSecond int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "problem in verifying the hash value with password", err)
		return
	}

	if !ok {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	var expirationTime time.Duration
	if params.ExpiresInSecond != 0 && params.ExpiresInSecond < int(time.Second*3600) {
		expirationTime = time.Duration(int(time.Second) * params.ExpiresInSecond)
	} else {
		expirationTime = time.Duration(time.Hour)
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong generating authentication token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})
}

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/priyanshoon/chirpy/internal/auth"
	"github.com/priyanshoon/chirpy/internal/database"
)

func (cfg *apiConfig) handlerPolka(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no header provided", err)
		return
	}

	if key != cfg.polkaKey {
		log.Println(key)
		log.Println(cfg.polkaKey)

		respondWithError(w, http.StatusUnauthorized,
			"you are not fucking authorized to make this fucking request", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "cannot decode the json", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userId, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "cannot parse the uuid", err)
		return
	}

	_, err = cfg.db.UpdateUserChirpyRed(r.Context(), database.UpdateUserChirpyRedParams{
		ID:          userId,
		IsChirpyRed: true,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// the book doesn't exist: client's problem
			respondWithError(w, http.StatusNotFound, "not found", err)
			return
		}
		// anything else: server's problem
		respondWithError(w, http.StatusInternalServerError,
			"couldn't update, something went wrong", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

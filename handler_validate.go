package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func validateChirpyHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChripLength = 140
	if len(params.Body) > maxChripLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedWords(params.Body, badWords)

	respondWithJSON(w, 200, returnVals{
		CleanedBody: cleaned,
	})
}

func getCleanedWords(body string, badWords map[string]struct{}) string {
	result := ""

	words := strings.Split(body, " ")
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if _, ok := badWords[lowerWord]; ok {
			words[i] = "****"
		}
	}

	result = strings.Join(words, " ")
	return result
}

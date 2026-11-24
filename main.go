package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/priyanshoon/chirpy/internal/database"
)

type apiConfig struct {
	db             *database.Queries
	platform       string
	fileserverHits atomic.Int32
	jwtSecret      string
	polkaKey       string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %s", err)
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("POLKA_KEY must be set")
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		db:             dbQueries,
		platform:       platform,
		fileserverHits: atomic.Int32{},
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
	}
	mux := http.NewServeMux()

	// health of server
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	// user related shit
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLoginUser)

	// chirps related shit
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpByID)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirpByID)

	// authentication handlers shit
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefersh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	// webhooks shit
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPolka)

	// this 2 should stay in for local test and development
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

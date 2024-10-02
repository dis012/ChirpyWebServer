package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/dis012/ChirpyWebServer/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	const port = ":8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	dbPlatform := os.Getenv("PLATFORM")
	if dbPlatform == "" {
		log.Fatal("PLATFORM environment variable is required")
	}

	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       dbPlatform,
		secret:         secret,
	}

	serverMux := http.NewServeMux()
	// Serve static files from the Chirpy/assets directory, stripping the /app prefix
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	// Wrap the file server with a middleware that increments the hit counter
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))
	// Register the /healthz endpoint for readiness checks
	serverMux.HandleFunc("GET /api/healthz", ReadinessHandler)
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	serverMux.HandleFunc("/admin/reset", apiCfg.resetMetricsHandler)
	serverMux.HandleFunc("POST /api/users", apiCfg.createNewUserHandler)
	serverMux.HandleFunc("POST /api/chirps", apiCfg.createNewChirpHandler)
	serverMux.HandleFunc("GET /api/chirps", apiCfg.getAllChirpsHandler)
	serverMux.HandleFunc("GET /api/chirps/{id}", apiCfg.getChirpByIdHandler)
	serverMux.HandleFunc("POST /api/login", apiCfg.loginUser)

	newServer := &http.Server{
		Addr:    port,
		Handler: serverMux,
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(newServer.ListenAndServe())
}

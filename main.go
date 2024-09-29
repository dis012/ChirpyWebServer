package main

import (
	"log"
	"net/http"
)

func main() {
	const port = ":8080"
	serverMux := http.NewServeMux()
	// Serve static files from the Chirpy/assets directory, stripping the /app prefix
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	// Wrap the file server with a middleware that increments the hit counter
	api := &apiConfig{}
	serverMux.Handle("/app/", api.middlewareMetricsInc(fileServer))
	// Register the /healthz endpoint for readiness checks
	serverMux.HandleFunc("GET /api/healthz", ReadinessHandler)
	serverMux.HandleFunc("GET /admin/metrics", api.metricsHandler)
	serverMux.HandleFunc("POST /admin/reset", api.reserMetricsHandler)

	newServer := &http.Server{
		Addr:    port,
		Handler: serverMux,
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(newServer.ListenAndServe())
}

package main

import "net/http"

// Readiness handler for /healthz
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Set the correct headers and body for the readiness endpoint
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK) // 200 OK
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

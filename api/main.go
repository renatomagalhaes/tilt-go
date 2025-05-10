package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Simple HTTP server that responds with "Hello, World!"
func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Define handler for root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World Tilt!")
	})

	// Start server
	log.Printf("Starting API server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

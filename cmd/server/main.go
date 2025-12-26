package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gremllm/schema/internal/middleware"
)

func main() {
	// Create a file server for the examples directory
	fs := http.FileServer(http.Dir("examples"))

	// Wrap it with our schema middleware
	handler := middleware.GremllmMiddleware(fs)

	// Start the server
	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Println("Try visiting:")
	fmt.Println("  - http://localhost:8080/index.html (renders HTML)")
	fmt.Println("  - http://localhost:8080/index.md (renders HTML for now)")

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}

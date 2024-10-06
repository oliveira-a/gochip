package main

import (
	"log"
	"net/http"
)

func main() {
	// Serve files from the current directory
	fs := http.FileServer(http.Dir("."))

	http.Handle("/", fs)

	log.Println("Serving on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

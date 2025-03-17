package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = ":8080"

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Printf("Listening on port %s...", port)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

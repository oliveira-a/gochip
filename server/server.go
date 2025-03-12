package main

import (
		"fmt"
		"log"
		"net/http"
)

func main() {
	fmt.Println("Listenning on port 8080...")

	http.Handle("/", http.FileServer(http.Dir(".")))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}


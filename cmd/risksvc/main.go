package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}
	fmt.Printf("Starting risksvc on port %s\n", port)
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "risksvc is healthy\n")
	})
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting risksvc: %v\n", err)
	}
}
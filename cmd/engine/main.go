package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	fmt.Printf("Starting engine on port %s\n", port)
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "engine is healthy\n")
	})
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting engine: %v\n", err)
	}
}
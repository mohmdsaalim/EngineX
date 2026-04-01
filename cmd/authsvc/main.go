package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Starting authsvc on port %s\n", port)
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "authsvc is healthy\n")
	})
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting authsvc: %v\n", err)
	}
}

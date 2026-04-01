package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	fmt.Printf("Starting executor on port %s\n", port)
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "executor is healthy\n")
	})
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting executor: %v\n", err)
	}
}
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"zenith/server/handlers"
)

func main() {
	var port string
	if len(os.Args) == 0 {
		port = ":8080"
		log.Printf("Using default port %s", port)
	} else {
		if _, err := strconv.Atoi(os.Args[1]); err != nil {
			log.Fatalf("Invalid port number: %s | %v", os.Args[1], err)
		}
		port = ":" + os.Args[1]
	}
	h := handlers.NewHandler()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /add", h.Add)
	mux.HandleFunc("POST /remove", h.Remove)
	mux.HandleFunc("GET /status", h.Status)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Printf("Listening on http://localhost%s", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

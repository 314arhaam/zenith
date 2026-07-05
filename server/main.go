package main

import (
	"log"
	"net/http"
	"zenith/server/handlers"
)

func main() {

	h := handlers.NewHandler()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /add", h.Add)
	mux.HandleFunc("POST /remove", h.Remove)
	mux.HandleFunc("GET /status", h.Status)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Listening on http://localhost%s", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

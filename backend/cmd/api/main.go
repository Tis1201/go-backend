package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Tis1201/go-backend/backend/internal/httpapi"
)

func main() {
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Listening on port %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

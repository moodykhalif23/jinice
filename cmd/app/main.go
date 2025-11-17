package main

import (
	"log"
	"net/http"
	"os"

	"example.com/starterkit/server"
)

func main() {
	// Initialize database
	if err := server.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	log.Println("Database initialized successfully")

	addr := ":8080"
	if a := os.Getenv("PORT"); a != "" {
		addr = ":" + a
	}
	mux := server.NewRouter()
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"
	"net/http"

	"simple-proxy/internal/config"
	"simple-proxy/internal/proxy"
)

func main() {
	cfg := config.Load()
	p := proxy.New()

	// Only handle CONNECT method for HTTPS tunneling
	handler := http.HandlerFunc(p.ServeHTTP)
	
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
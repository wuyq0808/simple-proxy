package main

import (
	"log"
	"net/http"

	"simple-proxy/internal/config"
	"simple-proxy/internal/proxy"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()
	p := proxy.New()

	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(p.ServeHTTP)
	
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
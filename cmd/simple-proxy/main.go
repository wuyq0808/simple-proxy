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

	// Create a custom handler that handles CONNECT before mux routing
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Handle CONNECT method directly (bypass mux)
		if req.Method == http.MethodConnect {
			p.ServeHTTP(w, req)
			return
		}
		
		// For all other methods, use mux router
		router := mux.NewRouter()
		router.PathPrefix("/").HandlerFunc(p.ServeHTTP)
		router.ServeHTTP(w, req)
	})
	
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
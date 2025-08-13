package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type Proxy struct{}

func New() *Proxy {
	return &Proxy{}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle HTTPS tunneling via CONNECT method
	if r.Method != "CONNECT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	p.handleConnect(w, r)
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	// Establish connection to target server
	targetConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to target: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// Send 200 Connection Established response
	w.WriteHeader(http.StatusOK)
	
	// Hijack the connection to get raw TCP access
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Failed to hijack connection: %v", err)
		return
	}
	defer clientConn.Close()

	// Bidirectional copy between client and target
	go func() {
		_, err := io.Copy(targetConn, clientConn)
		if err != nil {
			log.Printf("Client->Target copy error: %v", err)
		}
	}()

	_, err = io.Copy(clientConn, targetConn)
	if err != nil {
		log.Printf("Target->Client copy error: %v", err)
	}
}
package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Proxy struct {
	transport http.RoundTripper
}

func New() *Proxy {
	return &Proxy{
		transport: http.DefaultTransport,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== INCOMING REQUEST ===")
	log.Printf("Method: %s", r.Method)
	log.Printf("Host: %s", r.Host)
	log.Printf("URL: %s", r.URL.String())
	log.Printf("RequestURI: %s", r.URL.RequestURI())
	log.Printf("RemoteAddr: %s", r.RemoteAddr)

	// Handle HTTPS tunneling via CONNECT method
	if r.Method == "CONNECT" {
		p.handleConnect(w, r)
		return
	}

	// Handle regular HTTP requests (convert to HTTPS)
	target, err := url.Parse("https://" + r.Host + r.URL.RequestURI())
	if err != nil {
		log.Printf("URL Parse Error: %v", err)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	
	log.Printf("=== TARGET URL ===")
	log.Printf("Target: %s", target.String())
	log.Printf("Scheme: %s", target.Scheme)
	log.Printf("Host: %s", target.Host)
	log.Printf("Path: %s", target.Path)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			log.Printf("=== DIRECTOR ===")
			log.Printf("Before - URL: %s", req.URL.String())
			log.Printf("Before - Host: %s", req.Host)
			
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path
			req.URL.RawQuery = target.RawQuery
			req.Host = target.Host
			req.Header["X-Forwarded-For"] = nil
			
			log.Printf("After - URL: %s", req.URL.String())
			log.Printf("After - Host: %s", req.Host)
		},
		Transport: p.transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy Error: %v", err)
			http.Error(w, "Proxy Error", http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(w, r)
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== CONNECT REQUEST ===")
	log.Printf("Target Host: %s", r.Host)

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

	log.Printf("=== TUNNEL ESTABLISHED ===")
	log.Printf("Client: %s -> Proxy -> Target: %s", r.RemoteAddr, r.Host)

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

	log.Printf("=== TUNNEL CLOSED ===")
}
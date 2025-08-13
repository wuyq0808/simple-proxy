package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
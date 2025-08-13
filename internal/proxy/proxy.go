package proxy

import (
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
	target, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = target
			req.Host = target.Host
		},
		Transport: p.transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Proxy Error", http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(w, r)
}
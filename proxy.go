package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

type RelengapiProxy struct {
	listenPort     int
	targetHostname string
	permissions    []string
	issuingToken   string
}

func (rp RelengapiProxy) runForever() {
	// httputil's ReverseProxy is not specifically "reverse", and it will
	// do fine here.  The director transforms outgoing requests.
	director := func(req *http.Request) {
		// point toward the upstream server
		req.URL.Scheme = "https"
		req.URL.Host = rp.targetHostname
		req.Host = rp.targetHostname
		// Add the token
		req.Header.Add("Authorization", "Bearer 123")
		log.Println("Authenticating", req.URL)
	}
	proxy := &httputil.ReverseProxy{Director: director}

	// create a new HTTP server that handles everything via the proxy
	servemux := http.NewServeMux()
	servemux.HandleFunc("/", proxy.ServeHTTP)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", rp.listenPort), servemux))
}

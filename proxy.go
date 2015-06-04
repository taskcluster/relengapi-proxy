package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type RelengapiProxy struct {
	listenPort     int
	targetHostname string
	permissions    []string

	// token used to issue temporary tokens
	issuingToken string

	// temporary token and its expiration time
	tmpToken          string
	tmpTokenGoodUntil time.Time
}

// The temporary token is never exposed, so it doesn't have to have a
// super-short lifespan, but there's no sense in letting it remain valid for
// hours or days, as it will be renewed as necessary.
const tmpTokenLifetime time.Duration = 10 * time.Minute

// renew a little bit early to account for clock skew, etc.
const tmpTokenSkew time.Duration = 10 * time.Second

func (rp *RelengapiProxy) getToken() string {
	now := time.Now()
	if now.After(rp.tmpTokenGoodUntil) {
		expires := now.Add(tmpTokenLifetime)
		log.Printf("Generating new temporary token, expires at %v", expires)
		rp.tmpToken = getTmpToken(rp.targetHostname, rp.issuingToken, expires, rp.permissions)
		rp.tmpTokenGoodUntil = expires.Add(-tmpTokenSkew)
	}
	return rp.tmpToken
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
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rp.getToken()))
		// log
		log.Println(req.Method, req.URL)
	}
	proxy := &httputil.ReverseProxy{Director: director}

	// create a new HTTP server that handles everything via the proxy
	servemux := http.NewServeMux()
	servemux.HandleFunc("/", proxy.ServeHTTP)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", rp.listenPort), servemux))
}

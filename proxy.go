package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
	"strings"
)

type RelengapiProxy struct {
	listenPort    int
	relengapiHost string
	permissions   []string

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

func (rp *RelengapiProxy) getToken() (string, error) {
	now := time.Now()
	if now.After(rp.tmpTokenGoodUntil) {
		expires := now.Add(tmpTokenLifetime)
		log.Printf("Generating new temporary token; expires at %v", expires)
		tok, err := getTmpToken("https://tokens.mozilla-releng.net/tokens", rp.issuingToken, expires, rp.permissions)
		if err != nil {
			return "", err
		}
		rp.tmpToken = tok
		rp.tmpTokenGoodUntil = expires.Add(-tmpTokenSkew)
	}
	return rp.tmpToken, nil
}

func (rp RelengapiProxy) runForever() {
	log.Println("Proxying to RelengAPI with permissions:",
		rp.permissions, "on port", rp.listenPort)

	// httputil's ReverseProxy is not specifically "reverse", and it will
	// do fine here.  The director transforms outgoing requests.
	director := func(req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/tooltool") {
			req.URL.Scheme = "https"
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/tooltool")
			req.URL.RawPath = ""
			req.URL.Host = "tooltool.mozilla-releng.net"
			req.Host = "tooltool.mozilla-releng.net"
		} else if strings.HasPrefix(req.URL.Path, "/treestatus") {
			req.URL.Scheme = "https"
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/treestatus")
			req.URL.RawPath = ""
			req.URL.Host = "treestatus.mozilla-releng.net"
			req.Host = "treestatus.mozilla-releng.net"
		} else if strings.HasPrefix(req.URL.Path, "/mapper") {
			req.URL.Scheme = "https"
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/mapper")
			req.URL.RawPath = ""
			req.URL.Host = "mapper.mozilla-releng.net"
			req.Host = "mapper.mozilla-releng.net"
		} else {
			log.Fatal("Non exising service was requested.")
			return
		}
		// Add the token
		tok, err := rp.getToken()
		if err != nil {
			// ReverseProxy does not provide a way to short-circuit the
			// proxying and return an error response to the caller.  Anyway, if
			// we failed to get a token then the task is probably a complete
			// loss anyway.  So bail out.
			log.Fatal(err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))
		// log
		log.Println(req.Method, req.URL)

	}
	proxy := &httputil.ReverseProxy{Director: director}

	// create a new HTTP server that handles everything via the proxy
	servemux := http.NewServeMux()
	servemux.HandleFunc("/", proxy.ServeHTTP)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", rp.listenPort), servemux))
}

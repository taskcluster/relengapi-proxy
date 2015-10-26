package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func doScopesToPerms(t *testing.T, scopes, exp_perms []string) {
	perms := scopesToPerms(scopes)

	ok := len(perms) == len(exp_perms)
	if ok {
		for i, got := range perms {
			exp := exp_perms[i]
			if got != exp {
				ok = false
			}
		}
	}
	if !ok {
		t.Fatalf("scopesToPerms(%v) = %v; expected %v",
			scopes, perms, exp_perms)
	}
}

func TestScopesToPerms(t *testing.T) {
	// empty -> empty
	doScopesToPerms(t, []string{}, []string{})

	// ignore non-matching scopes
	doScopesToPerms(t, []string{
		"non-match:foo:bar",
		"docker-worker:relengapi-proxy:foo",
		"non-match2",
		"docker-worker:relengapi-proxy:bar",
	}, []string{
		"foo",
		"bar",
	})

	// no empty permissions
	doScopesToPerms(t, []string{
		"docker-worker:relengapi-proxy:",
	}, []string{})
}

func makeFakeServer(exp_expires time.Time, exp_perms []string, err_response bool) *httptest.Server {
	// a fake token-issuing endpoint that asserts a bunch of things about the
	// request
	serveHttp := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			log.Fatal("request method is not POST")
		}
		if r.URL.Path != "/tokenauth/tokens" {
			log.Fatal("request path is not /tokenauth/tokens")
		}
		if r.Header.Get("Authorization") != "Bearer iss-tok" {
			log.Fatalf("Authorization header not set correctly (%s)",
				r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			log.Fatalf("Content-Type header not set correctly (%s)",
				r.Header.Get("Content-Type"))
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		var reqBody interface{}
		err = json.Unmarshal(body, &reqBody)
		if err != nil {
			log.Fatal(err)
		}

		if reqBody.(map[string]interface{})["typ"].(string) != "tmp" {
			log.Fatalf("Bad typ in %s", body)
		}

		got_exp, err := time.Parse(time.RFC3339, reqBody.(map[string]interface{})["expires"].(string))
		if err != nil || got_exp != exp_expires {
			log.Fatalf("Bad expires in %s; expected %s but got %s",
				body, exp_expires, got_exp)
		}

		perms := reqBody.(map[string]interface{})["permissions"].([]interface{})
		if len(perms) != len(exp_perms) {
			log.Fatal("Did not get expected number of perms")
		}
		for i, _ := range perms {
			if perms[i] != exp_perms[i] {
				log.Fatalf("did not get correct permission %d", i)
			}
		}

		meta := reqBody.(map[string]interface{})["metadata"].(map[string]interface{})
		if len(meta) != 0 {
			log.Fatalf("Bad metadata in %s", body)
		}

		// build the response, eschewing json.Marshal and just supplying the text
		w.Header().Add("Content-Type", "application/json")
		if err_response {
			w.WriteHeader(500)
			fmt.Fprintf(w, `{
				"error": {
					"code": 500,
					"description": "BOOM"
				}
			}`)
		} else {
			fmt.Fprintf(w, `{
				"result": {
					"typ": "tmp",
					"token": "tmp-tok"
				}
			}`)
		}
	}

	// set up a fake relengapi server
	servemux := http.NewServeMux()
	servemux.HandleFunc("/tokenauth/tokens", serveHttp)
	return httptest.NewServer(servemux)
}

func TestGetTmpToken(t *testing.T) {
	expires := time.Date(2015, 6, 15, 13, 1, 1, 0, time.UTC)
	perms := []string{"perm.1", "perm.2"}

	ts := makeFakeServer(expires, perms, false)
	defer ts.Close()

	tok, err := getTmpToken(ts.URL, "iss-tok", expires, perms)

	if err != nil {
		t.Fatal(err)
	}
	if tok != "tmp-tok" {
		t.Fatalf("didn't get correct token")
	}
}

func TestGetTmpTokenFails(t *testing.T) {
	expires := time.Date(2015, 6, 15, 13, 1, 1, 0, time.UTC)
	perms := []string{"perm.1", "perm.2"}

	ts := makeFakeServer(expires, perms, true)
	defer ts.Close()

	tok, err := getTmpToken(ts.URL, "iss-tok", expires, perms)

	if err == nil {
		t.Fatal("didn't get error")
	}
	if tok != "" {
		t.Fatalf("token was not nil")
	}
}

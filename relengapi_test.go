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

func TestGetTmpToken(t *testing.T) {
	var expires time.Time

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
		if err != nil || got_exp != expires {
			log.Fatalf("Bad expires in %s; expected %s but got %s", body, expires, got_exp)
		}

		perms := reqBody.(map[string]interface{})["permissions"].([]interface{})
		if perms[0].(string) != "perm.1" {
			log.Fatalf("Bad permissions[0] in %s", body)
		}

		if perms[1].(string) != "perm.2" {
			log.Fatalf("Bad permissions[0] in %s", body)
		}

		meta := reqBody.(map[string]interface{})["metadata"].(map[string]interface{})
		if len(meta) != 0 {
			log.Fatalf("Bad metadata in %s", body)
		}

		// build the response, eschewing json.Marshal and just supplying the text
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"result": {
				"typ": "tmp",
				"token": "tmp-tok"
			}
		}`)
	}

	// set up a fake relengapi server
	servemux := http.NewServeMux()
	servemux.HandleFunc("/tokenauth/tokens", serveHttp)
	ts := httptest.NewServer(servemux)
	defer ts.Close()

	expires = time.Date(2015, 6, 15, 13, 1, 1, 0, time.UTC)
	getTmpToken(ts.URL, "iss-tok", expires, []string{"perm.1", "perm.2"})
}

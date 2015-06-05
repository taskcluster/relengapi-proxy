package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const ScopePrefix = "docker-worker:relengapi-proxy:"

func scopesToPerms(scopes []string) []string {
	var perms []string

	for _, scope := range scopes {
		if strings.HasPrefix(scope, ScopePrefix) {
			perm := scope[len(ScopePrefix):]
			if len(perm) != 0 {
				perms = append(perms, perm)
			}
		}
	}

	return perms
}

type relengapiTokenJson struct {
	Typ         string      `json:"typ"`
	Id          int         `json:"id,omitempty"`
	NotBefore   *time.Time  `json:"not_before,omitempty"`
	Expires     *time.Time  `json:"expires,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Permissions []string    `json:"permissions,omitempty"`
	Description string      `json:"description,omitempty"`
	User        string      `json:"user,omitempty"`
	Token       string      `json:"token,omitempty"`
}

func getTmpToken(url string, issuingToken string, expires time.Time, perms []string) string {
	// TODO: retry this operation
	request := relengapiTokenJson{
		Typ:         "tmp",
		Expires:     &expires,
		Permissions: perms,
		Metadata:    map[string]interface{}{},
	}

	reqbody, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	reqUrl := fmt.Sprintf("%s/tokenauth/tokens", url)
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(reqbody))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", issuingToken))
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Got '%s' while trying to get new tmp token:\n%s",
			resp.Status, string(body))
	}

	var responseBody interface{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		log.Fatal(err)
	}

	result := responseBody.(map[string]interface{})["result"]
	tok := result.(map[string]interface{})["token"].(string)
	return tok
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/taskcluster/httpbackoff"
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

func getTmpToken(urlPrefix string, issuingToken string, expires time.Time, perms []string) (tok string, err error) {
	request := relengapiTokenJson{
		Typ:         "tmp",
		Expires:     &expires,
		Permissions: perms,
		Metadata:    map[string]interface{}{},
	}

	reqbody, err := json.Marshal(request)
	if err != nil {
		return
	}

	client := &http.Client{}
	reqUrl := fmt.Sprintf("%s/tokenauth/tokens", urlPrefix)
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(reqbody))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", issuingToken))
	resp, _, err := httpbackoff.ClientDo(client, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf(
			"Got '%s' while trying to get new tmp token:\n%s",
			resp.Status, string(body)))
		return
	}

	var responseBody interface{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return
	}

	result := responseBody.(map[string]interface{})["result"]
	tok = result.(map[string]interface{})["token"].(string)
	return
}

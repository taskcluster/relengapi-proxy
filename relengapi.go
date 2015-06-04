package main

import (
	"strings"
)

func scopesToPerms(scopes []string) []string {
	var perms []string

	scopePrefix := "docker-worker:relengapi-proxy:"

	for _, scope := range scopes {
		if strings.HasPrefix(scope, scopePrefix) {
			perm := scope[len(scopePrefix):]
			if len(perm) != 0 {
				perms = append(perms, perm)
			}
		}
	}

	return perms
}

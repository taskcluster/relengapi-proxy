package main

import (
	"testing"
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

// TODO: test token fetching with httptest

package main

import (
	"log"
	"strconv"

	docopt "github.com/docopt/docopt-go"
)

var version = "RelengAPI proxy 1.0"
var usage = `
RelengAPI authentication proxy.

This attaches a temporary RelengAPI token to all outgoing requests to
RelengAPI.  The temporary token contains the permissions enumerated by task
scopes matching "docker-worker:relengapi-proxy:<perm>".  The temporary token is
generated via an HTTP request to RelengAPI using the permanent token given via
--relengapi-token, so any permissions not granted by that token cannot be
granted to a task.

  Usage:
    ./proxy [options] <taskId>
    ./proxy --help

  Options:
    -h --help                  Show this help screen.
    -p --port <port>           Port to bind the proxy server to [default: 8080].
    --relengapi-token <token>  RelengAPI token with which to reate temp tokens [default:].
	--relengapi-url <url>  	   RelengAPI URL [default: https://api.pub.build.mozilla.org].
`

func main() {
	arguments, err := docopt.Parse(usage, nil, true, version, false, true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	taskId := arguments["<taskId>"].(string)
	port, err := strconv.Atoi(arguments["--port"].(string))
	if err != nil {
		log.Fatalf("Failed to convert port to integer")
	}

	relengapiToken := arguments["--relengapi-token"].(string)
	if relengapiToken == "" {
		log.Fatal("--relengapi-token is required")
	}

	relengapiUrl := arguments["--relengapi-url"].(string)

	scopes, err := getTaskScopes(taskId)
	if err != nil {
		log.Fatalf("Could not fetch taskcluster task '%s' : %s", taskId, err)
	}
	relengapiPerms := scopesToPerms(scopes)

	if len(relengapiPerms) == 0 {
		log.Fatalf("No RelengAPI permission scopes (matching '%s*') found on task %s",
			ScopePrefix, taskId)
	}

	RelengapiProxy{
		listenPort:   port,
		relengapiUrl: relengapiUrl,
		permissions:  relengapiPerms,
		issuingToken: relengapiToken,
	}.runForever()
}

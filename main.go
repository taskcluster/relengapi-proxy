package main

import (
	//"fmt"
	"log"
	//"net/http"
	//"os"
	"strconv"

	docopt "github.com/docopt/docopt-go"
	// use `taskcluster` from the taskcluster proxy, for getTask
	tc "github.com/lightsofapollo/taskcluster-proxy/taskcluster"
)

var version = "RelengAPI proxy 1.0"
var usage = `
RelengAPI authentication proxy.

This attaches a temporary RelengAPI token to all outgoing requests to
RelengAPI.  The temporary token contains the permissions enumerated by scopes
matching "relengapi-proxy:permission:<perm>".  The temporary token is generated
via an HTTP request to RelengAPI using the permanent token given via
--relengapi-token, so any permissions not granted to that token cannot be granted
to a task.

  Usage:
    ./proxy [options] <taskId>
    ./proxy --help

  Options:
    -h --help                  Show this help screen.
    -p --port <port>           Port to bind the proxy server to [default: 8080].
    --relengapi-token <token>  The RelengAPI token with which to reate temp tokens
`

func main() {
	arguments, err := docopt.Parse(usage, nil, true, version, false, true)

	taskId := arguments["<taskId>"].(string)
	port, err := strconv.Atoi(arguments["--port"].(string))
	if err != nil {
		log.Fatalf("Failed to convert port to integer")
	}

	relengapiToken := arguments["--relengapi-token"]
	if relengapiToken == nil || relengapiToken == "" {
		log.Fatalf(
			"--relengapi-token is required",
		)
	}

	// Fetch the task to get the scopes we should be using...
	task, err := tc.GetTask(taskId)

	if err != nil {
		log.Fatalf("Could not fetch taskcluster task '%s' : %s", taskId, err)
	}

	log.Println("Proxy with scopes: ", task.Scopes, " on port ", port)
	/*
		routes := Routes{
			Scopes:      scopes,
			ClientId:    clientId.(string),
			AccessToken: relengapiToken.(string),
		}

		startError := http.ListenAndServe(fmt.Sprintf(":%d", port), routes)
		if startError != nil {
			log.Fatal(startError)
		}
	*/
}

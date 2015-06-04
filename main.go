package main

import (
	"log"
	"strconv"

	docopt "github.com/docopt/docopt-go"
	queue "github.com/taskcluster/taskcluster-client-go/queue"
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
    -h --help                        Show this help screen.
    -p --port <port>                 Port to bind the proxy server to [default: 8080].
    --relengapi-token <token>        The RelengAPI token with which to reate temp tokens [default:].
    --relengapi-hostname <hostname>  The RelengAPI hostname [default: api.pub.build.mozilla.org].
`

func main() {
	arguments, err := docopt.Parse(usage, nil, true, version, false, true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for k, v := range arguments {
		log.Println(k, v)
	}

	taskId := arguments["<taskId>"].(string)
	port, err := strconv.Atoi(arguments["--port"].(string))
	if err != nil {
		log.Fatalf("Failed to convert port to integer")
	}

	relengapiToken := arguments["--relengapi-token"].(string)
	if relengapiToken == "" {
		log.Fatalf(
			"--relengapi-token is required",
		)
	}

	relengapiHostname := arguments["--relengapi-hostname"].(string)

	// Fetch the task to get the scopes we should be using.  We don't need auth for this
	q := queue.New("", "")
	q.Authenticate = false
	task, callSummary := q.Task(taskId)
	if callSummary.Error != nil {
		log.Fatalf("Could not fetch taskcluster task '%s' : %s", taskId, callSummary.Error)
	}

	relengapiPerms := scopesToPerms(task.Scopes)

	log.Println("Proxy with scopes:", relengapiPerms, "on port", port)
	RelengapiProxy{
		listenPort:     port,
		targetHostname: relengapiHostname,
		permissions:    relengapiPerms,
		issuingToken:   relengapiToken,
	}.runForever()
}

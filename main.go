package main

import (
	"log"
	"strconv"

	docopt "github.com/docopt/docopt-go"
)

var version = "RelengAPI proxy 2.2.0"
var usage = `
RelengAPI authentication proxy.

This attaches a temporary RelengAPI token to all outgoing requests to
RelengAPI.  The temporary token contains the permissions enumerated by task
scopes matching "docker-worker:relengapi-proxy:<perm>".  The temporary token is
generated via an HTTP request to RelengAPI using the permanent token given via
--relengapi-token, so any permissions not granted by that token cannot be
granted to a task.

  Usage:
    relengapi-proxy [options] -- <taskId>
    relengapi-proxy -h|--help
    relengapi-proxy --version

  Options:
    -h --help                  Show this help screen.
    --version                  Show version.
    -p --port <port>           Port to bind the proxy server to [default: 8080].
    --relengapi-token <token>  RelengAPI token with which to create temp tokens [default:].
    --relengapi-host <url>     RelengAPI hostname [default: api.pub.build.mozilla.org].
`

func main() {
	arguments, err := parseProgramArgs(nil, true)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// treat nil (arg not provided) the same as an empty string (arg provided
	// but was empty string)
	fetchArg := func(argName string) (argValue string) {
		argValueInterface := arguments[argName]
		if argValueInterface != nil {
			argValue = argValueInterface.(string)
		}
		return argValue
	}

	taskId := fetchArg("<taskId>")
	if taskId == "" {
		log.Fatal("--task-id is required")
	}

	port, err := strconv.Atoi(fetchArg("--port"))
	if err != nil {
		log.Fatalf("Failed to convert port to integer (value supplied: %s)", arguments["--port"])
	}

	relengapiToken := fetchArg("--relengapi-token")
	if relengapiToken == "" {
		log.Fatal("--relengapi-token is required")
	}

	relengapiHost := fetchArg("--relengapi-host")

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
		listenPort:    port,
		relengapiHost: relengapiHost,
		permissions:   relengapiPerms,
		issuingToken:  relengapiToken,
	}.runForever()
}

func parseProgramArgs(argv []string, exit bool) (map[string]interface{}, error) {
	return docopt.Parse(usage, argv, true, version, false, exit)
}

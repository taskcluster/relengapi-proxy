# Taskcluster Proxy

This is the proxy server which is used in the docker-worker which allows
individual tasks to talk to various taskcluster services (auth, queue,
scheduler) without hardcoding credentials into the containers
themselves.

Credentials are expected to be passed via the `TASKCLUSTER_CLIENT_ID`
and `TASKCLUSTER_ACCESS_TOKEN` environment variables.


## Examples

Start the server, giving a relengapi token that can issue temporary tokens and a task ID

    relengapi-proxy --relengapi-token 12341234 2szAy1JzSr6pyjVCdiTcoQ

Once that's running, and assuming a docker alias mapping `relengapi:80` to the proxy,

    curl relengapi/tooltool/sha512/<some-sha512>

to download a file from tooltool, for example.

## Deployment

The proxy server can be deployed directly by building `proxy/main.go`
but the prefered method is via the `./build.sh` script which will
compile the proxy server for linux/amd64 and deploy the server to a
docker image.

```sh
./build.sh user/relengapi-proxy-server
```

## Download via `go get`

Set up your [GOPATH](https://golang.org/doc/code.html)

```sh
go get github.com/djmitche/relengapi-proxy
```

## Hacking

To build, just run

```sh
godep go build
```

## Tests

To run the full test suites you need a [RelengAPI](https://api.pub.build.mozilla.org/) token.
That token must have at least `base.tokens.tmp.issue`, as well as any permissions tasks may need.
The token is supplied with the --relengapi-token command-line argument.
Note that credentials must not be included in environment variables!

#! /bin/bash

set -e

help() {
  echo ""
  echo "Builds proxy server (For linux) and places into a docker container."
  echo "Docker and Go must be installed and able to compile linux/amd64."
  echo ""
  echo "  Usage: ./build.sh <docker image name>"
  echo ""
}

if [ -z "$1" ] ||
   [ "$1" == "-h" ] ||
   [ "$1" == "--help" ] ;
then
  help
  exit 0
fi

# step into directory of script
cd "$(dirname "${0}")"

GO_VERSION="$(go version 2>/dev/null | cut -f3 -d' ')"
GO_MAJ="$(echo "${GO_VERSION}" | cut -f1 -d'.')"
GO_MIN="$(echo "${GO_VERSION}" | cut -f2 -d'.')"
if [ -z "${GO_VERSION}" ]; then
  echo "Have you installed go? I get no result from \`go version\` command." >&2
  exit 64
elif [ "${GO_MAJ}" != "go1" ] || [ "${GO_MIN}" -lt 8 ]; then
  echo "Go version go1.x needed, where x >= 8, but the version I found is: '${GO_VERSION}'" >&2
  echo "I found it here:" >&2
  which go >&2
  echo "The complete output of \`go version\` command is:" >&2
  go version >&2
  exit 65
else
  echo "Go version ok! (${GO_VERSION})"
fi

uid="$(date +%s)"

# Output folder
mkdir -p target

echo "Fetching ca certs from latest ubuntu version..."
docker build --pull -t "${uid}" -f cacerts.docker .
docker run --name "${uid}" "${uid}"
docker cp "${uid}:/etc/ssl/certs/ca-certificates.crt" target
docker rm -v "${uid}"

echo "Building proxy server..."
GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o target/relengapi-proxy .

echo "Building docker image for proxy server"
docker build -t $1 .

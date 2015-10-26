FROM ubuntu:14.04
MAINTAINER Dustin J. Mitchell <dustin@mozilla.com>

RUN apt-get update
RUN apt-get install -y ca-certificates
EXPOSE 80
COPY target/relengapi-proxy /relengapi-proxy
ENTRYPOINT ["/relengapi-proxy", "--port", "80"]

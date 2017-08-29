FROM scratch
MAINTAINER Dustin J. Mitchell <dustin@mozilla.com>

EXPOSE 80
COPY target/relengapi-proxy /relengapi-proxy
COPY target/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/relengapi-proxy", "--port", "80"]

FROM quay.io/coreos/clair

RUN apk add --no-cache \
	ca-certificates

COPY testutils/snakeoil/cert.pem /usr/local/share/ca-certificates/clair.pem

# normally we'd use update-ca-certificates, but something about running it in
# Alpine is off, and the certs don't get added. Fortunately, we only need to
# add ca-certificates to the global store and it's all plain text.
RUN cat /usr/local/share/ca-certificates/* >> /etc/ssl/certs/ca-certificates.crt

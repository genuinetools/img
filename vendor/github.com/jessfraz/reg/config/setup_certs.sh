#!/bin/sh

CONFIGS_DIR=/etc/docker/daemon/config
CERT_DIR=/etc/docker/ssl

CERT_SUBJ="/C=US/ST=New York/L=New York City/O=Contained.AF/CN=Contained.AF CA"

if [ ! -f  "${CERT_DIR}/ca.pem" ]; then
	mkdir -p "${CERT_DIR}"

	# create the root CA
	openssl req -x509 \
		-config "${CONFIGS_DIR}/openssl-ca.cnf" \
		-newkey rsa:4096 -sha256 \
		-subj "${CERT_SUBJ}" \
		-nodes -out "${CERT_DIR}/ca.pem" -outform PEM

	openssl x509 -noout -text -in "${CERT_DIR}/ca.pem"

	# create the server certificate signing request
	openssl req \
		-config "${CONFIGS_DIR}/openssl-server.cnf" \
		-newkey rsa:2048 -sha256 \
		-subj "/CN=localhost" \
		-nodes -out "${CERT_DIR}/server.csr" -outform PEM
	openssl req -text -noout -verify -in "${CERT_DIR}/server.csr"

	touch "${CERT_DIR}/index.txt"
	echo 01 > "${CERT_DIR}/serial.txt"

	# create the server cert
	openssl ca -batch \
		-config "${CONFIGS_DIR}/openssl-ca.cnf" \
		-policy signing_policy -extensions signing_req \
		-out "${CERT_DIR}/cert.pem" -infiles "${CERT_DIR}/server.csr"

	openssl x509 -noout -text -in "${CERT_DIR}/cert.pem"

	# create the client certificate signing request
	openssl req \
		-config "${CONFIGS_DIR}/openssl-client.cnf" \
		-newkey rsa:2048 -sha256 \
		-subj "/CN=client" \
		-nodes -out "${CERT_DIR}/client.csr" -outform PEM
	openssl req -text -noout -verify -in "${CERT_DIR}/client.csr"

	touch "${CERT_DIR}/index.txt"
	echo 02 > "${CERT_DIR}/serial.txt"

	# create the client cert
	openssl ca -batch \
		-config "${CONFIGS_DIR}/openssl-ca.cnf" \
		-policy signing_policy -extensions signing_req \
		-out "${CERT_DIR}/client.cert" -infiles "${CERT_DIR}/client.csr"

	openssl x509 -noout -text -in "${CERT_DIR}/client.cert"


	# remove the signing requests
	rm -rf "${CERT_DIR}/client.csr" "${CERT_DIR}/server.csr" "${CERT_DIR}/"*.attr "${CERT_DIR}/"*.old

fi

set -- sh "$(which dind)" "$@"
exec "$@"

#!/bin/bash

CERT_DIR="./certs"

if [ -d $CERT_DIR ]; then
    echo "directiry is alredy exist"
else
    mkdir -p $CERT_DIR
fi

openssl genrsa -out $CERT_DIR/private.key 4096
openssl req -new -x509 -key $CERT_DIR/private.key -out $CERT_DIR/public.crt -days 365 -config ssl-haproxy.conf

cat public.crt private.key > haproxy-certificate.pem
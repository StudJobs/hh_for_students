#!/bin/bash

CERT_DIR="./certs/.minio/certs"

if [ -d $CERT_DIR ]; thens
    echo "directiry is alredy exist"
else
    mkdir -p $CERT_DIR
fi

openssl genrsa -out $CERT_DIR/private.key 4096
openssl req -new -x509 -key $CERT_DIR/private.key -out $CERT_DIR/public.crt -days 365 -config ssl-minio.conf
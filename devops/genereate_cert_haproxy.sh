#!/bin/sh

set -e

CERT_DIR="/tmp/certs"
CONFIG_FILE="/devops/ssl-haproxy.conf"

echo "Creating certificate directory: $CERT_DIR"
mkdir -p "$CERT_DIR"

echo "Generating private key..."
openssl genrsa -out "$CERT_DIR/private.key" 4096

echo "Generating certificate..."
if [ -f "$CONFIG_FILE" ]; then
    openssl req -new -x509 -key "$CERT_DIR/private.key" \
        -out "$CERT_DIR/public.crt" -days 365 \
        -config "$CONFIG_FILE"
else
    openssl req -new -x509 -key "$CERT_DIR/private.key" \
        -out "$CERT_DIR/public.crt" -days 365 \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
fi

echo "Combining certificate and key..."
cat "$CERT_DIR/public.crt" "$CERT_DIR/private.key" > "/usr/local/etc/haproxy/haproxy-certificate.pem"

# Создаем директорию для SSL и копируем туда
mkdir -p /etc/ssl/private
cp "/usr/local/etc/haproxy/haproxy-certificate.pem" "/etc/ssl/private/haproxy-certificate.pem"

echo "Certificate generation completed!"
echo "Certificate locations:"
ls -la "/usr/local/etc/haproxy/haproxy-certificate.pem"
ls -la "/etc/ssl/private/haproxy-certificate.pem"
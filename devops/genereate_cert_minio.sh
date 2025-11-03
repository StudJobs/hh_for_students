#!/bin/sh
# genereate_cert_minio.sh

set -e

CERT_DIR="/certs"
CONFIG_FILE="/certs/ssl-minio.conf"

echo "Creating certificate directory for MinIO..."
mkdir -p "$CERT_DIR"

echo "Generating MinIO private key..."
openssl genrsa -out "$CERT_DIR/private.key" 4096

echo "Generating MinIO certificate..."
if [ -f "$CONFIG_FILE" ]; then
    openssl req -new -x509 -key "$CERT_DIR/private.key" \
        -out "$CERT_DIR/public.crt" -days 365 \
        -config "$CONFIG_FILE"
else
    openssl req -new -x509 -key "$CERT_DIR/private.key" \
        -out "$CERT_DIR/public.crt" -days 365 \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=minio"
fi

# MinIO требует конкретные имена файлов
cp "$CERT_DIR/private.key" "$CERT_DIR/private.key"
cp "$CERT_DIR/public.crt" "$CERT_DIR/public.crt"

# Для MinIO также нужно скопировать в правильные пути
mkdir -p /root/.minio/certs
cp "$CERT_DIR/public.crt" /root/.minio/certs/public.crt
cp "$CERT_DIR/private.key" /root/.minio/certs/private.key

echo "MinIO certificates generated successfully!"
ls -la "$CERT_DIR/"
ls -la /root/.minio/certs/
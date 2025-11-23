#!/bin/bash
hostport=$1
timeout=${2:-60}

echo "Waiting for $hostport to be ready..."

for i in $(seq 1 $timeout); do
    if curl -s "http://$hostport" > /dev/null 2>&1; then
        echo "Service $hostport is ready!"
        exit 0
    fi
    echo "Waiting for $hostport... ($i/$timeout)"
    sleep 1
done

echo "Timeout waiting for $hostport"
exit 1
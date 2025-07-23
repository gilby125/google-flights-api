#!/bin/bash

# Script to generate self-signed TLS certificates for development
# DO NOT use these certificates in production!

set -e

# Create tls directory if it doesn't exist
mkdir -p tls

# Generate server certificate and key
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout tls/tls.key \
    -out tls/tls.crt \
    -subj "/C=US/ST=State/L=City/O=Development/CN=localhost"

# Generate CA certificate and key
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout tls/ca.key \
    -out tls/ca.crt \
    -subj "/C=US/ST=State/L=City/O=Development/CN=DevelopmentCA"

echo "TLS certificates generated successfully in tls/"
echo "These are self-signed certificates for development only!"
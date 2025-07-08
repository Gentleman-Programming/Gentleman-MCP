#!/bin/bash

# Generate development certificates for Gentleman MCP Gateway
# This creates a CA and server certificate for TLS development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_DIR="$SCRIPT_DIR/../certs"

echo "🔐 Generating development certificates for Gentleman MCP Gateway..."

# Create certs directory
mkdir -p "$CERTS_DIR"
cd "$CERTS_DIR"

# Generate CA private key
echo "1. Generating CA private key..."
openssl genrsa -out ca-key.pem 4096

# Generate CA certificate
echo "2. Generating CA certificate..."
openssl req -new -x509 -key ca-key.pem -sha256 -subj "/C=US/ST=CA/O=Gentleman MCP/CN=Gentleman MCP CA" -days 3650 -out ca-cert.pem

# Generate server private key
echo "3. Generating server private key..."
openssl genrsa -out server-key.pem 4096

# Generate server certificate signing request
echo "4. Generating server certificate signing request..."
openssl req -subj "/C=US/ST=CA/O=Gentleman MCP/CN=localhost" -sha256 -new -key server-key.pem -out server.csr

# Create extensions file for server certificate
cat > server-extfile.cnf <<EOF
subjectAltName = @alt_names
extendedKeyUsage = serverAuth

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

# Generate server certificate signed by CA
echo "5. Generating server certificate..."
openssl x509 -req -days 365 -sha256 -in server.csr -CA ca-cert.pem -CAkey ca-key.pem -out server-cert.pem -extfile server-extfile.cnf -CAcreateserial

# Generate client private key (for mTLS if needed)
echo "6. Generating client private key..."
openssl genrsa -out client-key.pem 4096

# Generate client certificate signing request
echo "7. Generating client certificate signing request..."
openssl req -subj "/C=US/ST=CA/O=Gentleman MCP/CN=client" -new -key client-key.pem -out client.csr

# Create extensions file for client certificate
cat > client-extfile.cnf <<EOF
extendedKeyUsage = clientAuth
EOF

# Generate client certificate signed by CA
echo "8. Generating client certificate..."
openssl x509 -req -days 365 -sha256 -in client.csr -CA ca-cert.pem -CAkey ca-key.pem -out client-cert.pem -extfile client-extfile.cnf -CAcreateserial

# Clean up temporary files
rm server.csr client.csr server-extfile.cnf client-extfile.cnf

# Set appropriate permissions
chmod 400 *-key.pem
chmod 444 *-cert.pem ca-cert.pem

echo "✅ Certificates generated successfully!"
echo ""
echo "Generated files:"
echo "  📁 $CERTS_DIR/"
echo "    🔑 ca-key.pem       - CA private key"
echo "    📜 ca-cert.pem      - CA certificate"
echo "    🔑 server-key.pem   - Server private key"
echo "    📜 server-cert.pem  - Server certificate"
echo "    🔑 client-key.pem   - Client private key (for mTLS)"
echo "    📜 client-cert.pem  - Client certificate (for mTLS)"
echo ""
echo "🚀 Ready for TLS development!"
echo "   Server will use: server-cert.pem + server-key.pem"
echo "   Client will use: client-cert.pem + client-key.pem (optional mTLS)"
echo "   CA certificate: ca-cert.pem"

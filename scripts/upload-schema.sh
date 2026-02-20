#!/bin/bash
# Upload schema to ISR (Internal Schema Registry)
# Usage: ./scripts/upload-schema.sh <version>

set -e

VERSION="${1:-1.0.0}"
SCHEMA_FILE="/tmp/schema-descriptor.bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "📦 Building schema descriptor..."
buf build -o "$SCHEMA_FILE"

# Build and run the upload client
cd "$SCRIPT_DIR/upload-client"
go run main.go "$VERSION" "$SCHEMA_FILE"

# Cleanup
rm -f "$SCHEMA_FILE"

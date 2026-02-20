#!/bin/bash
# Pull schema from ISR (Internal Schema Registry)
# Usage: ./scripts/pull-schema.sh <major> <minor>

set -e

MAJOR="${1:-1}"
MINOR="${2:-0}"
ISR_URL="${CELO_ISR_URL:-localhost:50051}"
OUTPUT_FILE="/tmp/schema-descriptor-pulled.bin"

echo "📥 Pulling schema for version $MAJOR.$MINOR from ISR ($ISR_URL)..."

RESPONSE=$(curl -s -X POST "http://$ISR_URL/isr.v1.SchemaRegistryService/GetLatestPatch" \
  -H "Content-Type: application/json" \
  -d "{
    \"major\": $MAJOR,
    \"minor\": $MINOR
  }")

# Check if response contains error
if echo "$RESPONSE" | grep -q "error\|code"; then
  echo "❌ Pull failed:"
  echo "$RESPONSE" | jq '.'
  exit 1
fi

# Extract metadata
SCHEMA_ID=$(echo "$RESPONSE" | jq -r '.metadata.id')
VERSION=$(echo "$RESPONSE" | jq -r '.metadata.version')
SIZE=$(echo "$RESPONSE" | jq -r '.metadata.sizeBytes')

echo "✅ Schema found!"
echo "  ID: $SCHEMA_ID"
echo "  Version: $VERSION"
echo "  Size: $SIZE bytes"

# Extract and decode binary
SCHEMA_BASE64=$(echo "$RESPONSE" | jq -r '.schemaBinary')
echo "$SCHEMA_BASE64" | base64 -d > "$OUTPUT_FILE"

echo ""
echo "💾 Schema saved to: $OUTPUT_FILE"
echo "📝 You can inspect it with: buf ls-files $OUTPUT_FILE"

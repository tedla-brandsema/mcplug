#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CONFIG_PATH="examples/ngrok-dev/mcpfs.cfg.json"

if [[ -z "${NGROK_AUTHTOKEN:-}" ]]; then
  echo "ERROR: NGROK_AUTHTOKEN is not set."
  echo
  echo "Usage:"
  echo "  NGROK_AUTHTOKEN='<your-ngrok-token>' ./scripts/run-ngrok.sh"
  echo
  exit 1
fi

if [[ -z "${MCPFS_TOKEN:-}" ]]; then
  export MCPFS_TOKEN="$(openssl rand -hex 32)"
fi

echo "Running tests..."
go test ./...

echo "Building mcpfs..."
mkdir -p ./bin
go build -o ./bin/mcpfs ./cmd/mcpfs

echo
echo "MCPFS_TOKEN:"
echo "$MCPFS_TOKEN"
echo
echo "Starting mcpfs with embedded ngrok..."
echo "Config: $CONFIG_PATH"
echo

exec ./bin/mcpfs -config "$CONFIG_PATH"

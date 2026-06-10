#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CONFIG_PATH="examples/ngrok-dev/mcplug.cfg.json"

if [[ -z "${NGROK_AUTHTOKEN:-}" ]]; then
  echo "ERROR: NGROK_AUTHTOKEN is not set."
  echo
  echo "Usage:"
  echo "  NGROK_AUTHTOKEN='<your-ngrok-token>' ./scripts/run-ngrok.sh"
  echo
  exit 1
fi

if [[ -z "${MCPLUG_TOKEN:-}" ]]; then
  export MCPLUG_TOKEN="$(openssl rand -hex 32)"
fi

echo "Running tests..."
go test ./...

echo "Building plug..."
mkdir -p ./bin
go build -o ./bin/plug ./cmd/plug

echo
echo "MCPLUG_TOKEN:"
echo "$MCPLUG_TOKEN"
echo
echo "Starting plug with embedded ngrok..."
echo "Config: $CONFIG_PATH"
echo

exec ./bin/plug -config "$CONFIG_PATH"

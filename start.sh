#!/usr/bin/env bash

set -euo pipefail

trap 'kill -- -$$' EXIT

docker compose up -d

OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
  go run ./cmd/connect-playground-server
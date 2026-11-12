#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="${HOME}/go/bin:/usr/local/go/bin:${PATH}"
export HTTP_ADDR="${HTTP_ADDR:-:18080}"
export DATABASE_URL="${DATABASE_URL:-fusion:123456@tcp(127.0.0.1:43306)/ai_fusion_video?parseTime=true&loc=Local}"
export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:46379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-123456}"
exec go run ./cmd/server

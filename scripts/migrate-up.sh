#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="${HOME}/go/bin:/usr/local/go/bin:${PATH}"
export DATABASE_URL="${DATABASE_URL:-fusion:123456@tcp(127.0.0.1:43306)/ai_fusion_video?parseTime=true&loc=Local}"
goose -dir db/migrations mysql "$DATABASE_URL" up
goose -dir db/migrations mysql "$DATABASE_URL" status

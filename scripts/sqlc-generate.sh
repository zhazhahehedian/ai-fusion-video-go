#!/usr/bin/env bash
# sqlc 1.30 的 MySQL 函数目录只有零参数 UTC_TIMESTAMP，无法类型检查 UTC_TIMESTAMP(6)。
# 生成时临时换成返回类型相同的 CURRENT_TIMESTAMP(6)，再把任务 SQL 写回 UTC。
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

queries=(db/queries/generation.sql db/queries/media.sql)
generated=(internal/sqlcdb/generation.sql.go internal/sqlcdb/media.sql.go)

restore() {
  local f
  for f in "${queries[@]}"; do
    if [[ -f "$f.bak-sqlc" ]]; then
      mv "$f.bak-sqlc" "$f"
    fi
  done
}
trap restore EXIT

for f in "${queries[@]}"; do
  cp "$f" "$f.bak-sqlc"
  python3 - "$f" << 'PY'
import pathlib, sys
path = pathlib.Path(sys.argv[1])
path.write_text(path.read_text().replace("UTC_TIMESTAMP(6)", "CURRENT_TIMESTAMP(6)"))
PY
done

sqlc generate -f db/sqlc.yaml

for f in "${generated[@]}"; do
  python3 - "$f" << 'PY'
import pathlib, sys
path = pathlib.Path(sys.argv[1])
path.write_text(path.read_text().replace("CURRENT_TIMESTAMP(6)", "UTC_TIMESTAMP(6)"))
PY
done

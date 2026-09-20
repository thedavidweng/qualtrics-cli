#!/usr/bin/env bash
# Endpoint catalog drift check: every endpoint marked done in docs/endpoints.yaml
# must have its path referenced in non-test code, and vice versa.
set -euo pipefail

cd "$(dirname "$0")/.."

catalog="docs/endpoints.yaml"
if [[ ! -f "$catalog" ]]; then
  echo "missing $catalog"
  exit 1
fi

code_files=$(find internal -name '*.go' ! -name '*_test.go')
missing=0

while IFS= read -r route; do
  [[ -z "$route" ]] && continue
  if ! grep -qhF "\"$route" $code_files; then
    echo "drift: $route is marked done in $catalog but no code references it"
    missing=1
  fi
done < <(python3 - "$catalog" <<'PY'
import sys, yaml
doc = yaml.safe_load(open(sys.argv[1]))
roots = set()
for ep in doc.get("endpoints", []):
    if ep.get("status") == "done":
        parts = ep["path"].strip("/").split("/")
        roots.add("/" + parts[0])
for r in sorted(roots):
    print(r)
PY
)

while IFS= read -r path; do
  [[ -z "$path" ]] && continue
  first="${path#/}"; first="${first%%/*}"
  if ! grep -q "^  - path: /$first\$" "$catalog" && ! grep -q "^  - path: /$first/" "$catalog"; then
    continue
  fi
  if ! grep -qF "path: $path" "$catalog"; then
    echo "drift: code references $path but it is absent from $catalog"
    missing=1
  fi
done < <(grep -hoE '"(/[a-zA-Z0-9/_{}.-]+)"' $code_files | tr -d '"' | sort -u)

if [[ "$missing" == "1" ]]; then
  exit 1
fi
echo "endpoint catalog in sync"

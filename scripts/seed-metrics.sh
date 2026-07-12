#!/usr/bin/env bash
# Seed bulk metric data for dashboard verification.
# Usage: bash scripts/seed-metrics.sh
set -euo pipefail

API="${RECENSION_API_URL:-http://localhost:8080}"
KEY="${RECENSION_API_KEY:-dev-api-key}"
PASS="${RECENSION_BOOTSTRAP_PASSWORD:-dev-password}"
TEAM="${RECENSION_TEAM:-acme}"
SUITE="${RECENSION_SUITE:-students}"
BASE_REV="${BASE_REV:-metrics-base-$(date +%H%M%S)}"
CMP_REV="${CMP_REV:-metrics-cmp-$(date +%H%M%S)}"

NAMES=(alice bob charlie diana eve frank grace henry ivy jack
  karen leo mia noah olivia paul quinn rose sam tara)

metric_for() {
  local i="$1" rev="$2"
  local find_ms=$((8 + (i % 7) * 3))
  local render_ms=$((20 + (i % 5) * 4))
  if [[ "$rev" == "$CMP_REV" ]]; then
    case $((i % 4)) in
      0) find_ms=$((find_ms + 12)) ;;
      1) render_ms=$((render_ms + 25)) ;;
      2) find_ms=$((find_ms - 2)); if [[ $find_ms -lt 1 ]]; then find_ms=1; fi ;;
      3) ;;
    esac
  fi
  printf '%s' "$find_ms $render_ms"
}

submit_batch() {
  local rev="$1"
  local include_extra="$2"
  local now
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  local msgs="["
  local i=0
  local first=1
  for name in "${NAMES[@]}"; do
    read -r find_ms render_ms <<<"$(metric_for "$i" "$rev")"
    local metrics
    metrics=$(cat <<EOF
[{"key":"find_student","value":${find_ms}},{"key":"render_card","value":${render_ms}}]
EOF
)
    if [[ "$include_extra" == "1" && $((i % 5)) -eq 0 ]]; then
      metrics=$(cat <<EOF
[{"key":"find_student","value":${find_ms}},{"key":"render_card","value":${render_ms}},{"key":"cache_hit","value":$((1 + i))}]
EOF
)
    fi
    if [[ "$rev" == "$CMP_REV" && $((i % 6)) -eq 0 ]]; then
      metrics=$(cat <<EOF
[{"key":"find_student","value":${find_ms}}]
EOF
)
      if [[ "$include_extra" == "1" ]]; then
        metrics=$(cat <<EOF
[{"key":"find_student","value":${find_ms}},{"key":"cache_hit","value":$((1 + i))}]
EOF
)
      fi
    fi

    local full
    full="$(python3 -c "print('$name'.capitalize() + ' Example')")"
    local gpa
    gpa=$(python3 -c "print(round(2.5 + ($i % 15) * 0.1, 1))")
    local chunk
    chunk=$(cat <<EOF
{"metadata":{"team":"$TEAM","suite":"$SUITE","version":"$rev","testcase":"$name","builtAt":"$now"},"results":[{"key":"username","kind":"assert","value":{"type":"string","string":"$name"}},{"key":"fullname","kind":"check","value":{"type":"string","string":"$full"}},{"key":"gpa","kind":"check","value":{"type":"double","double":$gpa}}],"metrics":$metrics}
EOF
)
    if [[ $first -eq 1 ]]; then
      msgs+="$chunk"
      first=0
    else
      msgs+=",$chunk"
    fi
    i=$((i + 1))
  done
  msgs+="]"

  echo "==> submit $TEAM/$SUITE/$rev (${#NAMES[@]} testcases)"
  local code body
  body="$(mktemp)"
  code="$(curl -sS -o "$body" -w '%{http_code}' -X POST "$API/v1/client/submit" \
    -H "X-Recension-API-Key: $KEY" \
    -H 'Content-Type: application/json' \
    -d "{\"version\":1,\"messages\":$msgs}")"
  if [[ "$code" != "200" ]]; then
    echo "submit failed HTTP $code: $(tr '\n' ' ' <"$body")" >&2
    rm -f "$body"
    exit 1
  fi
  rm -f "$body"

  echo "==> seal $rev"
  code="$(curl -sS -o /dev/null -w '%{http_code}' -X POST \
    "$API/v1/batch/$TEAM/$SUITE/$rev/seal" \
    -H "X-Recension-API-Key: $KEY")"
  if [[ "$code" != "204" && "$code" != "200" ]]; then
    echo "seal failed HTTP $code" >&2
    exit 1
  fi
}

echo "API=$API team=$TEAM suite=$SUITE"
curl -sf "$API/healthz" >/dev/null

submit_batch "$BASE_REV" 0

echo "==> promote $BASE_REV as baseline"
TOKEN="$(curl -sS -X POST "$API/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"dev@recension.local\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"
code="$(curl -sS -o /dev/null -w '%{http_code}' -X POST \
  "$API/v1/batch/$TEAM/$SUITE/$BASE_REV/promote" \
  -H "Authorization: Bearer $TOKEN")"
if [[ "$code" != "204" ]]; then
  echo "promote failed HTTP $code" >&2
  exit 1
fi

submit_batch "$CMP_REV" 1

echo
echo "Seeded."
echo "  Baseline: $BASE_REV  (20 cases — find_student + render_card)"
echo "  Compare:  $CMP_REV   (diffs / fresh cache_hit / missing render_card)"
echo
echo "Dashboard:"
echo "  http://localhost:3000/t/$TEAM/$SUITE/$CMP_REV"
echo "  http://localhost:3000/t/$TEAM/$SUITE/$CMP_REV/e/alice   # slower find + fresh cache_hit + missing render"
echo "  http://localhost:3000/t/$TEAM/$SUITE/$CMP_REV/e/bob     # slower render"
echo "  http://localhost:3000/t/$TEAM/$SUITE/$CMP_REV/e/diana   # unchanged metrics"

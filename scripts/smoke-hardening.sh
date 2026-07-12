#!/usr/bin/env bash
# Smoke-test Recension hardening against a temporary API process.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
API_DIR="$ROOT/api"
PORT="${SMOKE_PORT:-18080}"
BASE="http://127.0.0.1:${PORT}"
KEY="${RECENSION_API_KEY:-dev-api-key}"
PASS="${RECENSION_BOOTSTRAP_PASSWORD:-dev-password}"
DB_URL="${RECENSION_DATABASE_URL:-postgres://recension:recension@localhost:5432/recension?sslmode=disable}"
BIN="${TMPDIR:-/tmp}/recension-smoke-api"
TEAM="acme"
SUITE="students"
BATCH="smoke-$(date +%s)"
API_PID=""
PASS_COUNT=0
FAIL_COUNT=0

cleanup() {
  stop_api
  rm -f "$BIN"
}
trap cleanup EXIT

stop_api() {
  if [[ -n "${API_PID:-}" ]] && kill -0 "$API_PID" 2>/dev/null; then
    kill "$API_PID" 2>/dev/null || true
    wait "$API_PID" 2>/dev/null || true
  fi
  API_PID=""
  # Extra safety: anything still bound to the smoke port.
  if command -v lsof >/dev/null 2>&1; then
    local pids
    pids="$(lsof -tiTCP:"$PORT" -sTCP:LISTEN 2>/dev/null || true)"
    if [[ -n "$pids" ]]; then
      kill $pids 2>/dev/null || true
      sleep 0.2
    fi
  fi
}

wait_down() {
  local i
  for i in $(seq 1 40); do
    if ! curl -sf "$BASE/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.25
  done
  return 1
}

wait_up() {
  local i
  for i in $(seq 1 60); do
    if curl -sf "$BASE/healthz" >/dev/null 2>&1; then
      return 0
    fi
    if [[ -n "$API_PID" ]] && ! kill -0 "$API_PID" 2>/dev/null; then
      return 1
    fi
    sleep 0.25
  done
  return 1
}

start_api() {
  # Usage: start_api ALLOW_SIGNUP BOOTSTRAP
  local allow_signup="$1"
  local bootstrap="$2"
  stop_api
  wait_down || true
  (
    RECENSION_ENV=development \
    RECENSION_PORT="$PORT" \
    RECENSION_API_KEY="$KEY" \
    RECENSION_BOOTSTRAP_PASSWORD="$PASS" \
    RECENSION_SESSION_SECRET=recension-dev-secret-change-me \
    RECENSION_BOOTSTRAP="$bootstrap" \
    RECENSION_ALLOW_SIGNUP="$allow_signup" \
    RECENSION_DATABASE_URL="$DB_URL" \
    RECENSION_CORS_ORIGINS="*" \
    "$BIN"
  ) >/tmp/recension-smoke-api.log 2>&1 &
  API_PID=$!
  if ! wait_up; then
    fail "API did not become healthy (log: $(tr '\n' ' ' </tmp/recension-smoke-api.log))"
    echo "PASS=$PASS_COUNT FAIL=$FAIL_COUNT"
    exit 1
  fi
}

ok() {
  PASS_COUNT=$((PASS_COUNT + 1))
  printf '  PASS  %s\n' "$1"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  printf '  FAIL  %s\n' "$1"
}

expect_status() {
  local name="$1" want="$2" method="$3" url="$4"
  shift 4
  local code body
  body="$(mktemp)"
  code="$(curl -sS -o "$body" -w '%{http_code}' -X "$method" "$url" "$@" || true)"
  if [[ "$code" == "$want" ]]; then
    ok "$name (HTTP $code)"
  else
    fail "$name (want $want, got $code: $(tr '\n' ' ' <"$body"))"
  fi
  rm -f "$body"
}

echo "==> ensuring postgres"
docker compose -f "$ROOT/ops/compose.yaml" up -d postgres >/dev/null

echo "==> building API binary"
(cd "$API_DIR" && go build -o "$BIN" ./cmd/server)

echo "==> production config must refuse defaults"
if (
  RECENSION_ENV=production \
  RECENSION_DATABASE_URL="$DB_URL" \
  "$BIN"
) >/tmp/recension-smoke-prod.log 2>&1; then
  fail "production start with defaults should exit non-zero"
else
  if grep -qiE 'non-default|config:' /tmp/recension-smoke-prod.log; then
    ok "production rejects default secrets"
  else
    fail "production failed for unexpected reason: $(tr '\n' ' ' </tmp/recension-smoke-prod.log)"
  fi
fi

echo "==> starting API on :$PORT (signup on)"
start_api true true
ok "healthz"

echo "==> auth"
expect_status "login bootstrap" 200 POST "$BASE/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"dev@recension.local\",\"password\":\"$PASS\"}"

TOKEN="$(curl -sS -X POST "$BASE/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"dev@recension.local\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

SMOKE_EMAIL="smoke-$(date +%s)@recension.local"
SMOKE_SIGNUP="$(curl -sS -w '\n%{http_code}' -X POST "$BASE/v1/auth/signup" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$SMOKE_EMAIL\",\"password\":\"password1\"}")"
SMOKE_SIGNUP_BODY="$(printf '%s' "$SMOKE_SIGNUP" | sed '$d')"
SMOKE_SIGNUP_CODE="$(printf '%s' "$SMOKE_SIGNUP" | tail -n1)"
if [[ "$SMOKE_SIGNUP_CODE" == "200" ]]; then
  ok "signup allowed (HTTP 200)"
else
  fail "signup allowed (want 200, got $SMOKE_SIGNUP_CODE: $SMOKE_SIGNUP_BODY)"
fi
SMOKE_KEY="$(printf '%s' "$SMOKE_SIGNUP_BODY" | python3 -c 'import sys,json; print(json.load(sys.stdin)["user"]["apiKey"])')"
SMOKE_TOKEN="$(printf '%s' "$SMOKE_SIGNUP_BODY" | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

echo "==> submit hardening"
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
MSG_OK=$(cat <<EOF
{"messages":[{"metadata":{"team":"$TEAM","suite":"$SUITE","version":"$BATCH","testcase":"alice","builtAt":"$NOW"},"results":[{"key":"fullname","kind":"check","value":{"type":"string","string":"Alice Anderson"}}],"metrics":[]}]}
EOF
)
MSG_BAD_TEAM=$(cat <<EOF
{"messages":[{"metadata":{"team":"no-such-team","suite":"$SUITE","version":"$BATCH","testcase":"alice","builtAt":"$NOW"},"results":[],"metrics":[]}]}
EOF
)
MSG_BAD_SUITE=$(cat <<EOF
{"messages":[{"metadata":{"team":"$TEAM","suite":"no-such-suite","version":"$BATCH","testcase":"alice","builtAt":"$NOW"},"results":[],"metrics":[]}]}
EOF
)

expect_status "unknown team rejected" 404 POST "$BASE/v1/client/submit" \
  -H "X-Recension-API-Key: $KEY" -H 'Content-Type: application/json' \
  -d "$MSG_BAD_TEAM"

expect_status "unknown suite rejected" 404 POST "$BASE/v1/client/submit" \
  -H "X-Recension-API-Key: $KEY" -H 'Content-Type: application/json' \
  -d "$MSG_BAD_SUITE"

expect_status "submit ok" 200 POST "$BASE/v1/client/submit" \
  -H "X-Recension-API-Key: $KEY" -H 'Content-Type: application/json' \
  -d "$MSG_OK"

expect_status "seal batch" 204 POST "$BASE/v1/batch/$TEAM/$SUITE/$BATCH/seal" \
  -H "X-Recension-API-Key: $KEY"

expect_status "sealed submit rejected" 409 POST "$BASE/v1/client/submit" \
  -H "X-Recension-API-Key: $KEY" -H 'Content-Type: application/json' \
  -d "$MSG_OK"

echo "==> headers / body limit"
HDRS="$(curl -sSI "$BASE/healthz")"
if echo "$HDRS" | grep -qi 'X-Content-Type-Options: nosniff'; then
  ok "security header nosniff"
else
  fail "missing X-Content-Type-Options"
fi

expect_status "bad api key" 401 POST "$BASE/v1/client/submit" \
  -H 'X-Recension-API-Key: wrong-key' -H 'Content-Type: application/json' \
  -d "$MSG_OK"

echo "==> dashboard authz still works"
expect_status "list batches with session" 200 GET "$BASE/v1/teams/$TEAM/suites/$SUITE/batches" \
  -H "Authorization: Bearer $TOKEN"

echo "==> tenant isolation"
expect_status "outsider cannot list acme suites" 404 GET "$BASE/v1/teams/$TEAM/suites" \
  -H "Authorization: Bearer $SMOKE_TOKEN"

expect_status "outsider cannot promote" 404 POST "$BASE/v1/batch/$TEAM/$SUITE/$BATCH/promote" \
  -H "Authorization: Bearer $SMOKE_TOKEN"

expect_status "outsider key cannot submit to acme" 404 POST "$BASE/v1/client/submit" \
  -H "X-Recension-API-Key: $SMOKE_KEY" -H 'Content-Type: application/json' \
  -d "$MSG_OK"

expect_status "outsider teams list empty" 200 GET "$BASE/v1/teams" \
  -H "Authorization: Bearer $SMOKE_TOKEN"

TEAMS_LEN="$(curl -sS "$BASE/v1/teams" -H "Authorization: Bearer $SMOKE_TOKEN" \
  | python3 -c 'import sys,json; print(len(json.load(sys.stdin)))')"
if [[ "$TEAMS_LEN" == "0" ]]; then
  ok "outsider sees zero teams"
else
  fail "outsider should see 0 teams, got $TEAMS_LEN"
fi

echo "==> restart with signup disabled"
start_api false false
ok "signup-disabled API up"

expect_status "signup forbidden" 403 POST "$BASE/v1/auth/signup" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"blocked-$(date +%s)@recension.local\",\"password\":\"password1\"}"

expect_status "login still works" 200 POST "$BASE/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"dev@recension.local\",\"password\":\"$PASS\"}"

echo
echo "PASS=$PASS_COUNT FAIL=$FAIL_COUNT"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  exit 1
fi
echo "smoke-hardening ok"

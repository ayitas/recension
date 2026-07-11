#!/usr/bin/env bash
# Smoke-test MinIO blob upload / head / get + SDK []byte path.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
API_DIR="$ROOT/api"
PORT="${SMOKE_PORT:-18081}"
BASE="http://127.0.0.1:${PORT}"
KEY="${RECENSION_API_KEY:-dev-api-key}"
PASS="${RECENSION_BOOTSTRAP_PASSWORD:-dev-password}"
DB_URL="${RECENSION_DATABASE_URL:-postgres://recension:recension@localhost:5432/recension?sslmode=disable}"
BIN="${TMPDIR:-/tmp}/recension-smoke-blobs-api"
API_PID=""
PASS_COUNT=0
FAIL_COUNT=0

cleanup() {
  if [[ -n "${API_PID:-}" ]] && kill -0 "$API_PID" 2>/dev/null; then
    kill "$API_PID" 2>/dev/null || true
    wait "$API_PID" 2>/dev/null || true
  fi
  if command -v lsof >/dev/null 2>&1; then
    pids="$(lsof -tiTCP:"$PORT" -sTCP:LISTEN 2>/dev/null || true)"
    [[ -n "$pids" ]] && kill $pids 2>/dev/null || true
  fi
  rm -f "$BIN"
}
trap cleanup EXIT

ok() { PASS_COUNT=$((PASS_COUNT + 1)); printf '  PASS  %s\n' "$1"; }
fail() { FAIL_COUNT=$((FAIL_COUNT + 1)); printf '  FAIL  %s\n' "$1"; }

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

echo "==> ensuring postgres + minio"
docker compose -f "$ROOT/ops/compose.yaml" up -d postgres minio >/dev/null

echo "==> waiting for minio"
for _ in $(seq 1 40); do
  if curl -sf http://127.0.0.1:9000/minio/health/live >/dev/null 2>&1; then
    break
  fi
  sleep 0.25
done

echo "==> building API"
(cd "$API_DIR" && go build -o "$BIN" ./cmd/server)

echo "==> starting API on :$PORT"
(
  RECENSION_ENV=development \
  RECENSION_PORT="$PORT" \
  RECENSION_API_KEY="$KEY" \
  RECENSION_BOOTSTRAP_PASSWORD="$PASS" \
  RECENSION_SESSION_SECRET=recension-dev-secret-change-me \
  RECENSION_BOOTSTRAP=true \
  RECENSION_ALLOW_SIGNUP=true \
  RECENSION_DATABASE_URL="$DB_URL" \
  RECENSION_S3_ENDPOINT=localhost:9000 \
  RECENSION_S3_ACCESS_KEY=recension \
  RECENSION_S3_SECRET_KEY=recensionsecret \
  RECENSION_S3_BUCKET=recension \
  RECENSION_S3_USE_SSL=false \
  "$BIN"
) >/tmp/recension-smoke-blobs.log 2>&1 &
API_PID=$!

for _ in $(seq 1 60); do
  if curl -sf "$BASE/healthz" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "$API_PID" 2>/dev/null; then
    fail "API exited: $(tr '\n' ' ' </tmp/recension-smoke-blobs.log)"
    echo "PASS=$PASS_COUNT FAIL=$FAIL_COUNT"
    exit 1
  fi
  sleep 0.25
done
ok "healthz"

PAYLOAD="hello-recension-blob-$RANDOM"
DIGEST="sha256:$(printf '%s' "$PAYLOAD" | shasum -a 256 | awk '{print $1}')"
# Go ServeMux PathValue unescapes; PathEscape encodes colon.
DIG_PATH="$(python3 -c "import urllib.parse; print(urllib.parse.quote('$DIGEST', safe=''))")"

expect_status "head missing" 404 HEAD "$BASE/v1/blobs/$DIG_PATH" \
  -H "X-Recension-API-Key: $KEY"

expect_status "put blob" 200 PUT "$BASE/v1/blobs/$DIG_PATH" \
  -H "X-Recension-API-Key: $KEY" \
  -H "Content-Type: text/plain" \
  --data-binary "$PAYLOAD"

expect_status "head exists" 204 HEAD "$BASE/v1/blobs/$DIG_PATH" \
  -H "X-Recension-API-Key: $KEY"

GOT="$(curl -sS "$BASE/v1/blobs/$DIG_PATH" -H "X-Recension-API-Key: $KEY")"
if [[ "$GOT" == "$PAYLOAD" ]]; then
  ok "get blob bytes match"
else
  fail "get blob mismatch: $GOT"
fi

# digest mismatch
BAD="sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
BAD_PATH="$(python3 -c "import urllib.parse; print(urllib.parse.quote('$BAD', safe=''))")"
expect_status "put digest mismatch" 400 PUT "$BASE/v1/blobs/$BAD_PATH" \
  -H "X-Recension-API-Key: $KEY" \
  --data-binary "$PAYLOAD"

# submit with blob ref
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
BATCH="blob-smoke-$(date +%s)"
MSG=$(cat <<EOF
{"messages":[{"metadata":{"team":"acme","suite":"students","version":"$BATCH","testcase":"alice","builtAt":"$NOW"},"results":[{"key":"artifact","kind":"check","value":{"type":"blob","blob":{"digest":"$DIGEST","mime":"text/plain"}}}],"metrics":[]}]}
EOF
)
expect_status "submit with blob ref" 200 POST "$BASE/v1/client/submit" \
  -H "X-Recension-API-Key: $KEY" -H 'Content-Type: application/json' \
  -d "$MSG"

echo
echo "PASS=$PASS_COUNT FAIL=$FAIL_COUNT"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  echo "api log: $(tr '\n' ' ' </tmp/recension-smoke-blobs.log)"
  exit 1
fi
echo "smoke-blobs ok"

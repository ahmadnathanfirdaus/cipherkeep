#!/usr/bin/env bash
#
# End-to-end smoke test for the cipherkeep API.
# Exercises the core flow: health -> register -> login -> create project ->
# create environment -> create secret -> read secret (verify) -> refresh token.
#
# Usage:
#   API_BASE_URL=http://localhost:8080/api/v1 ./scripts/smoke-test.sh
#
# Requires: curl, python3. Exits non-zero on the first failed assertion.

set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080/api/v1}"
HEALTH_URL="${HEALTH_URL:-http://localhost:8080/healthz}"

# Unique email per run so repeated runs don't collide.
EMAIL="smoke+$(date +%s)@example.com"
PASSWORD="SuperSecret123!"
NAME="Smoke Test"

pass() { printf '  \033[32mPASS\033[0m %s\n' "$1"; }
fail() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; exit 1; }

# json <json-string> <python-expression-on-variable d>
json() { python3 -c 'import sys,json; d=json.load(sys.stdin); print(eval(sys.argv[1]))' "$1"; }

echo "==> API: $API_BASE_URL"

echo "[1/8] Health check"
curl -fsS "$HEALTH_URL" >/dev/null && pass "healthz reachable" || fail "healthz unreachable"

echo "[2/8] Register"
curl -fsS -X POST "$API_BASE_URL/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"name\":\"$NAME\",\"password\":\"$PASSWORD\"}" >/dev/null \
  && pass "registered $EMAIL" || fail "register failed"

echo "[3/8] Login"
LOGIN=$(curl -fsS -X POST "$API_BASE_URL/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
ACCESS=$(printf '%s' "$LOGIN" | json "d['data']['access_token']")
REFRESH=$(printf '%s' "$LOGIN" | json "d['data']['refresh_token']")
[ -n "$ACCESS" ] && pass "got access token" || fail "no access token"

AUTH=(-H "Authorization: Bearer $ACCESS")

echo "[4/8] Create project"
PROJECT=$(curl -fsS -X POST "$API_BASE_URL/projects" "${AUTH[@]}" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Smoke Project","description":"created by smoke test"}')
PROJECT_ID=$(printf '%s' "$PROJECT" | json "d['data']['project']['id']")
[ -n "$PROJECT_ID" ] && pass "project $PROJECT_ID" || fail "create project failed"

echo "[5/8] Create environment"
ENVIRONMENT=$(curl -fsS -X POST "$API_BASE_URL/projects/$PROJECT_ID/environments" "${AUTH[@]}" \
  -H 'Content-Type: application/json' \
  -d '{"name":"production"}')
ENV_ID=$(printf '%s' "$ENVIRONMENT" | json "d['data']['environment']['id']")
[ -n "$ENV_ID" ] && pass "environment $ENV_ID" || fail "create environment failed"

echo "[6/8] Create secret"
SECRET_VALUE="postgres://user:pass@db:5432/app"
curl -fsS -X POST "$API_BASE_URL/environments/$ENV_ID/secrets" "${AUTH[@]}" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"DATABASE_URL\",\"value\":\"$SECRET_VALUE\"}" >/dev/null \
  && pass "secret DATABASE_URL created" || fail "create secret failed"

echo "[7/8] Read secret and verify decryption round-trip"
READ=$(curl -fsS "$API_BASE_URL/environments/$ENV_ID/secrets/DATABASE_URL" "${AUTH[@]}")
GOT=$(printf '%s' "$READ" | json "d['data']['secret']['value']")
if [ "$GOT" = "$SECRET_VALUE" ]; then
  pass "decrypted value matches original"
else
  fail "value mismatch: got '$GOT'"
fi

echo "[8/8] Refresh token rotation"
REFRESHED=$(curl -fsS -X POST "$API_BASE_URL/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}")
NEW_ACCESS=$(printf '%s' "$REFRESHED" | json "d['data']['access_token']")
[ -n "$NEW_ACCESS" ] && pass "refreshed access token" || fail "refresh failed"

printf '\n\033[32mAll smoke tests passed.\033[0m\n'

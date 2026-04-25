#!/bin/sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

if [ -d /migrations ]; then
  i=0
  ok=0
  while [ "$i" -lt 30 ]; do
    if migrate -path /migrations -database "$DATABASE_URL" up; then
      ok=1
      break
    fi
    i=$((i + 1))
    echo "migrate failed, retrying ($i/30)..." >&2
    sleep 1
  done
  if [ "$ok" -ne 1 ]; then
    echo "migrate failed after retries" >&2
    exit 1
  fi
fi

exec /usr/local/bin/api

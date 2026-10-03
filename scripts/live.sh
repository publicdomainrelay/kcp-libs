#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "usage: live.sh '<command to run against a shared kcp>'" >&2
  exit 2
fi

root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
cluster_pid=""

cleanup() {
  if [ -n "$cluster_pid" ]; then
    kill "$cluster_pid" 2>/dev/null || true
    wait "$cluster_pid" 2>/dev/null || true
  fi
  rm -rf "$tmp"
}
trap cleanup EXIT

go build -o "$tmp/livecluster" "$root/internal/livekcp/cmd/livecluster"
"$tmp/livecluster" --env "$tmp/env" >"$tmp/out" 2>"$tmp/log" &
cluster_pid=$!

for _ in $(seq 1 900); do
  [ -s "$tmp/env" ] && break
  kill -0 "$cluster_pid" 2>/dev/null || break
  sleep 0.2
done

if [ ! -s "$tmp/env" ]; then
  echo "the shared kcp did not start:" >&2
  cat "$tmp/log" >&2
  exit 1
fi

# shellcheck disable=SC1091
set -a
. "$tmp/env"
set +a
export KCP_LIBS_REQUIRE_LIVE=1

cd "$root"
bash -c "$1"

#!/bin/zsh
set -e

cd /Users/chenrui/Desktop/go-migration-platform

for pidfile in .runlogs/*.pid; do
  [ -f "$pidfile" ] || continue
  pid=$(cat "$pidfile")
  kill "$pid" 2>/dev/null || true
  rm -f "$pidfile"
done

echo "stopped local go services"

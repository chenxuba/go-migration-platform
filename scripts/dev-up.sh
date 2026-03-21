#!/bin/zsh
set -e

cd /Users/chenrui/Desktop/go-migration-platform

mkdir -p .runlogs

~/.local/go/current/bin/go run ./services/iam/cmd/api > .runlogs/iam.log 2>&1 &
echo $! > .runlogs/iam.pid

~/.local/go/current/bin/go run ./services/platform/cmd/api > .runlogs/platform.log 2>&1 &
echo $! > .runlogs/platform.pid

~/.local/go/current/bin/go run ./services/education/cmd/api > .runlogs/education.log 2>&1 &
echo $! > .runlogs/education.pid

echo "iam-service      http://127.0.0.1:8081"
echo "platform-service http://127.0.0.1:8082"
echo "education-service http://127.0.0.1:8083"
echo "logs in .runlogs/"

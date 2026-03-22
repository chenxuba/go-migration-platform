#!/bin/zsh
set -e

ROOT="${0:A:h}/.."
cd "${ROOT:A}"

GO="${GO:-$HOME/.local/go/current/bin/go}"
if [[ ! -x "$GO" ]]; then
  GO=$(command -v go) || { echo "找不到 go，请设置 PATH 或环境变量 GO=/path/to/go"; exit 1; }
fi

mkdir -p .runlogs

echo "[启动 1/3] iam-service → 后台 go run，日志 .runlogs/iam.log"
"$GO" run ./services/iam/cmd/api > .runlogs/iam.log 2>&1 &
echo $! > .runlogs/iam.pid

echo "[启动 2/3] platform-service → .runlogs/platform.log"
"$GO" run ./services/platform/cmd/api > .runlogs/platform.log 2>&1 &
echo $! > .runlogs/platform.pid

echo "[启动 3/3] education-service → .runlogs/education.log（首次编译常最慢）"
"$GO" run ./services/education/cmd/api > .runlogs/education.log 2>&1 &
echo $! > .runlogs/education.pid

echo "iam-service       http://127.0.0.1:8081"
echo "platform-service  http://127.0.0.1:8082"
echo "education-service http://127.0.0.1:8083"
echo "logs in .runlogs/"

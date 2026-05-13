#!/bin/zsh
# 仅启动 Go；不检查 NATS/Meilisearch。完整流程（起中间件 + 预检 + 本脚本）请用 scripts/restart.sh
set -e

ROOT="${0:A:h}/.."
cd "${ROOT:A}"

GO="${GO:-$HOME/.local/go/current/bin/go}"
if [[ ! -x "$GO" ]]; then
  GO=$(command -v go) || { echo "找不到 go，请设置 PATH 或环境变量 GO=/path/to/go"; exit 1; }
fi

mkdir -p .runlogs
if [[ -f .runlogs/platform-ai.env ]]; then
  set -a
  source .runlogs/platform-ai.env
  set +a
fi

ensure_port_free() {
  local port="$1"
  local service="$2"
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "端口 ${port} 已被占用，无法启动 ${service}。当前监听进程：" >&2
    lsof -nP -iTCP:"$port" -sTCP:LISTEN >&2 || true
    echo "请先执行 scripts/dev-down.sh，或手动停止上面的进程。" >&2
    exit 1
  fi
}

echo "==> 检查 8081/8082/8083 是否可用…"
ensure_port_free 8081 iam-service
ensure_port_free 8082 platform-service
ensure_port_free 8083 education-service
rm -f .runlogs/iam.pid .runlogs/platform.pid .runlogs/education.pid

echo "[启动 1/3] iam-service → 后台 go run，日志 .runlogs/iam.log"
nohup "$GO" run ./services/iam/cmd/api > .runlogs/iam.log 2>&1 &
echo $! > .runlogs/iam.pid

echo "[启动 2/3] platform-service → .runlogs/platform.log"
nohup "$GO" run ./services/platform/cmd/api > .runlogs/platform.log 2>&1 &
echo $! > .runlogs/platform.pid

echo "[启动 3/3] education-service → .runlogs/education.log（首次编译常最慢）"
nohup "$GO" run ./services/education/cmd/api > .runlogs/education.log 2>&1 &
echo $! > .runlogs/education.pid

echo "iam-service       http://127.0.0.1:8081"
echo "platform-service  http://127.0.0.1:8082"
echo "education-service http://127.0.0.1:8083"
echo "logs in .runlogs/"

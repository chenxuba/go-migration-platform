#!/bin/zsh
# 重启：ensure-dev-infra（按需起 MQ/ES/Canal，参考 Java start_all_services）→ preflight → dev-down → dev-up → 等 8081–8083。
# 依赖未起时直接退出，不拉 Go（SKIP_PREFLIGHT=1 跳过预检；SKIP_ENSURE_INFRA=1 跳过自动起中间件）。
#
# 用法: ./scripts/restart.sh
#   或: zsh /path/to/go-migration-platform/scripts/restart.sh

set -e

SCRIPT_DIR="${0:A:h}"
cd "${SCRIPT_DIR:A}/.."
cd "${PWD:A}"

echo "==> 1/5 按需拉起本地中间件（RocketMQ / ES / Canal，见 scripts/ensure-dev-infra.sh）…"
zsh "${SCRIPT_DIR}/ensure-dev-infra.sh" || {
  echo "中间件未就绪，已中止。可检查 ROCKETMQ_HOME、CANAL_HOME，或 SKIP_ENSURE_INFRA=1 后自行启动。" >&2
  exit 1
}

echo "==> 2/5 依赖预检（RocketMQ + ES + Canal）…"
zsh "${SCRIPT_DIR}/preflight-dev-deps.sh" || {
  echo "依赖未就绪，已中止（不启动 Go）。可 SKIP_PREFLIGHT=1 跳过检查。" >&2
  exit 1
}

echo "==> 3/5 停止旧进程..."
zsh "${SCRIPT_DIR}/dev-down.sh"

echo "==> 4/5 启动新进程（后台，日志在 .runlogs/）..."
zsh "${SCRIPT_DIR}/dev-up.sh"

echo ""
echo "已写入 PID（当前 shell 启动的 go run；编译完成前端口可能尚未监听）:"
for f in .runlogs/*.pid(N); do
  [[ -f "$f" ]] || continue
  echo "  ${f:t:r}: $(<"$f")"
done

echo ""
echo "==> 5/5 等待端口 8081(iam) / 8082(platform) / 8083(education) 监听…"
echo "（无实时编译百分比：go 输出在各自 .log 里；下面每 2 秒刷一次状态）"
echo ""

typeset -i waited=0
typeset -i last_report=-999

port_listening() {
  lsof -iTCP:"$1" -sTCP:LISTEN >/dev/null 2>&1
}

while (( waited < 120 )); do
  typeset -i n=0
  typeset line=""
  for p in 8081 8082 8083; do
    if port_listening "$p"; then
      n=$((n + 1))
      line+=" ${p}:OK"
    else
      line+=" ${p}:…"
    fi
  done
  (( n == 3 )) && {
    echo "  就绪，共耗时 ${waited}s —$line"
    break
  }

  if (( waited == 0 || waited - last_report >= 2 )); then
    printf '  [%3ds]%s\n' "$waited" "$line"
    last_report=waited
  fi

  sleep 1
  waited=$((waited + 1))
done

typeset -i ready_count=0
for p in 8081 8082 8083; do
  port_listening "$p" && ready_count=$((ready_count + 1))
done
if (( ready_count < 3 && waited >= 120 )); then
  echo ""
  echo "  已等满 120s 仍未全部监听。另开终端可看实时编译输出:"
  echo "    tail -f .runlogs/education.log"
fi

echo ""
for p in 8081 8082 8083; do
  if lsof -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "  端口 $p: 已在监听"
  else
    echo "  端口 $p: 仍未监听 — 请 tail -f .runlogs/*.log 看是否编译报错"
  fi
done
echo ""
echo "完成。要确认是否换过进程，可对比上面的 PID 与重启前是否不同。"

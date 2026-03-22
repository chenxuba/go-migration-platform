#!/bin/zsh
# 停止本仓库对应的本地 Go 服务。
# 仅 kill .pid 往往不够：go run 监听端口的可能是子进程，或进程根本不是通过 dev-up 起的。
# 因此会再按端口 8081 / 8082 / 8083 清理 LISTEN 进程。

ROOT="${0:A:h}/.."
cd "${ROOT:A}"

mkdir -p .runlogs

echo "==> 1) 按 .runlogs/*.pid 发送 SIGTERM…"
for pidfile in .runlogs/*.pid(N); do
  [[ -f "$pidfile" ]] || continue
  pid=$(tr -d '[:space:]' <"$pidfile" | head -1)
  [[ -n "$pid" ]] || continue
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid" 2>/dev/null && echo "    pid $pid ($pidfile:t)"
  fi
  rm -f "$pidfile"
done

echo "==> 2) 按端口 8081 8082 8083 结束监听进程（SIGTERM）…"
for port in 8081 8082 8083; do
  pids=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null | sort -u)
  for p in ${(f)pids}; do
    [[ -n "$p" ]] || continue
    kill "$p" 2>/dev/null && echo "    pid $p (port $port)"
  done
done

sleep 1

echo "==> 3) 仍在监听的端口 → SIGKILL…"
for port in 8081 8082 8083; do
  pids=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null | sort -u)
  for p in ${(f)pids}; do
    [[ -n "$p" ]] || continue
    kill -9 "$p" 2>/dev/null && echo "    KILL pid $p (port $port)"
  done
done

echo ""
typeset -i left=0
for port in 8081 8082 8083; do
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "警告: 端口 $port 仍有监听，请执行: lsof -nP -iTCP:$port -sTCP:LISTEN"
    left=1
  fi
done

if (( left == 0 )); then
  echo "完成: 8081/8082/8083 已无监听。"
  echo "说明: 若前端页面仍能「拉到数据」，多半是代理指向了别的地址（例如远程/Java/其它端口），与本次停止的本地 Go 无关。"
else
  echo "完成: 部分端口未释放，请按上行 lsof 手动处理。"
fi

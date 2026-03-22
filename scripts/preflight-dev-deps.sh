#!/bin/zsh
# 启动 education 前：检查 RocketMQ、ES（TCP）、Canal（与 Java start_all_services 一致：优先看 canal.deployer 进程）。
# 只做「端口/进程」，不要求配 ES 账号。
# 环境变量与 pkg/config/config.go 中 Load 使用的名称一致。
#
#   ROCKETMQ_NAMESRV       默认 127.0.0.1:9876（多个用 ; 时检查第一个）
#   ES_URI                 默认 https://127.0.0.1:9200（只解析 host:port 做 TCP）
#   Canal                  默认必检：进程 canal.deployer；若还设置了 CANAL_HEALTH_URL，则「进程或该端口 TCP」任一满足即通过
#   PREFLIGHT_SKIP_CANAL=1 不检查 Canal（纯 Go / 无 CDC 时用）
#   SKIP_PREFLIGHT=1       跳过本脚本全部检查
#   PREFLIGHT_TIMEOUT      每项最长等待秒数，默认 120
#
# 若仍要做 ES 集群状态 + 账号校验，可另开终端手动：
#   curl -sk -u "$ES_USERNAME:$ES_PASSWORD" "$ES_URI/_cluster/health"

[[ -n "$SKIP_PREFLIGHT" ]] && exit 0

typeset -i TIMEOUT=${PREFLIGHT_TIMEOUT:-120}

tcp_open() {
  nc -z -w 2 "$1" "$2" >/dev/null 2>&1
}

wait_tcp() {
  local label="$1" host="$2" port="$3"
  typeset -i t=0
  echo "  等待 $label (${host}:${port}) …"
  while (( t < TIMEOUT )); do
    tcp_open "$host" "$port" && {
      echo "    $label 已就绪"
      return 0
    }
    (( t % 5 == 0 && t > 0 )) && echo "    … 已等待 ${t}s"
    sleep 1
    t=$((t + 1))
  done
  echo "    错误: ${TIMEOUT}s 内 $label 仍不可连" >&2
  return 1
}

# 从 http(s)://host:port/path 解析 host 与 port（缺省端口时由 default_port 兜底）
parse_http_host_port() {
  local raw="$1" default_port="$2"
  local u="${raw#http://}"
  u="${u#https://}"
  local hp="${u%%/*}"
  local h="${hp%%:*}"
  local p="${hp##*:}"
  [[ "$p" == "$h" ]] && p="$default_port"
  print -r -- "$h" "$p"
}

echo "==> 依赖预检（RocketMQ + ES + Canal）每项最长 ${TIMEOUT}s"
echo "    跳过全部: SKIP_PREFLIGHT=1  |  不检 Canal: PREFLIGHT_SKIP_CANAL=1"

# --- RocketMQ NameServer ---
mq="${ROCKETMQ_NAMESRV:-127.0.0.1:9876}"
mq="${mq%%;*}"
mq="${mq#http://}"
mq="${mq#https://}"
mhost="${mq%%:*}"
mport="${mq##*:}"
[[ "$mport" == "$mhost" ]] && mport=9876
wait_tcp "RocketMQ NameServer" "$mhost" "$mport" || exit 1

# --- Elasticsearch：只确认监听端口（不调用 HTTP，避免 xpack 未带账号时出现 401）---
es="${ES_URI:-https://127.0.0.1:9200}"
read -r es_host es_port <<<"$(parse_http_host_port "$es" 9200)"
wait_tcp "Elasticsearch（HTTP 端口）" "$es_host" "$es_port" || exit 1

# --- Canal：与 Java 脚本一致，以 canal.deployer 进程为准；可选再验管理/健康 URL 端口 ---
canal_ok() {
  pgrep -f 'canal.deployer' >/dev/null 2>&1 && return 0
  if [[ -n "${CANAL_HEALTH_URL:-}" ]]; then
    local ch cp
    read -r ch cp <<<"$(parse_http_host_port "$CANAL_HEALTH_URL" 8089)"
    tcp_open "$ch" "$cp" && return 0
  fi
  return 1
}

wait_canal() {
  typeset -i t=0
  echo "  等待 Canal（进程 canal.deployer${CANAL_HEALTH_URL:+ 或 $CANAL_HEALTH_URL TCP}）…"
  while (( t < TIMEOUT )); do
    canal_ok && {
      echo "    Canal 已就绪"
      return 0
    }
    (( t % 5 == 0 && t > 0 )) && echo "    … 已等待 ${t}s"
    sleep 1
    t=$((t + 1))
  done
  echo "    错误: ${TIMEOUT}s 内 Canal 仍不可用（请确认已运行 scripts/ensure-dev-infra.sh 或手动 startup.sh）" >&2
  return 1
}

if [[ -n "${PREFLIGHT_SKIP_CANAL:-}" ]]; then
  echo "  跳过 Canal（PREFLIGHT_SKIP_CANAL）"
else
  wait_canal || exit 1
fi

echo "==> 依赖预检通过"
exit 0

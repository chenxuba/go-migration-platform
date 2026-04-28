#!/bin/zsh
# 启动 education 前：检查 NATS JetStream、Meilisearch 是否可连。
#
#   NATS_URL             默认 nats://127.0.0.1:4222
#   MEILI_HOST           默认 http://127.0.0.1:7700
#   SKIP_PREFLIGHT=1     跳过本脚本全部检查
#   PREFLIGHT_TIMEOUT    每项最长等待秒数，默认 120

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

parse_nats_host_port() {
  local raw="$1"
  raw="${raw%%,*}"
  raw="${raw#nats://}"
  raw="${raw#tls://}"
  raw="${raw#ws://}"
  raw="${raw#wss://}"
  raw="${raw#http://}"
  raw="${raw#https://}"
  raw="${raw#*@}"
  local hp="${raw%%/*}"
  local h="${hp%%:*}"
  local p="${hp##*:}"
  [[ "$p" == "$h" ]] && p="4222"
  print -r -- "$h" "$p"
}

echo "==> 依赖预检（NATS JetStream + Meilisearch）每项最长 ${TIMEOUT}s"
echo "    跳过全部: SKIP_PREFLIGHT=1"

nats_url="${NATS_URL:-nats://127.0.0.1:4222}"
read -r nats_host nats_port <<<"$(parse_nats_host_port "$nats_url")"
wait_tcp "NATS JetStream" "$nats_host" "$nats_port" || exit 1

meili_host="${MEILI_HOST:-http://127.0.0.1:7700}"
read -r meili_addr meili_port <<<"$(parse_http_host_port "$meili_host" 7700)"
wait_tcp "Meilisearch" "$meili_addr" "$meili_port" || exit 1

echo "==> 依赖预检通过"
exit 0

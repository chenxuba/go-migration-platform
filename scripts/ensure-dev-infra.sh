#!/bin/zsh
# 本地开发：若 RocketMQ（NameServer + Broker）、Elasticsearch、Canal 未就绪则尝试拉起。
# 逻辑参考 5.0-project-ai/start_all_services.sh（brew ES + ROCKETMQ_HOME + CANAL_HOME）。
#
#   SKIP_ENSURE_INFRA=1     跳过本脚本（CI / 仅检查时用）
#   ROCKETMQ_HOME           默认 ~/rocketmq
#   CANAL_HOME              默认 /usr/local/canal.deployer-1.1.8
#   ROCKETMQ_NAMESRV        与 config 一致，默认 127.0.0.1:9876（Broker 启动用 -n）
#   ROCKETMQ_BROKER_HOST    默认 127.0.0.1（仅做 TCP 探活）
#   ROCKETMQ_BROKER_PORT    默认 10911
#   ES_URI                  仅解析 host:port；默认对 127.0.0.1:9200 探活并尝试 brew 启动
#   ENSURE_INFRA_TIMEOUT    每项最长等待秒数，默认 120

[[ -n "$SKIP_ENSURE_INFRA" ]] && {
  echo "  [ensure-dev-infra] 已跳过（SKIP_ENSURE_INFRA）"
  exit 0
}

typeset -i TIMEOUT=${ENSURE_INFRA_TIMEOUT:-120}

tcp_open() {
  nc -z -w 2 "$1" "$2" >/dev/null 2>&1
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

wait_tcp() {
  local label="$1" host="$2" port="$3"
  typeset -i t=0
  echo "  [ensure-dev-infra] 等待 $label (${host}:${port}) …"
  while (( t < TIMEOUT )); do
    tcp_open "$host" "$port" && {
      echo "  [ensure-dev-infra] $label 已就绪"
      return 0
    }
    (( t % 5 == 0 && t > 0 )) && echo "  [ensure-dev-infra] … 已等待 ${t}s"
    sleep 1
    t=$((t + 1))
  done
  echo "  [ensure-dev-infra] 错误: ${TIMEOUT}s 内 $label 仍不可连" >&2
  return 1
}

canal_process_running() {
  pgrep -f 'canal.deployer' >/dev/null 2>&1
}

wait_canal_process() {
  typeset -i t=0
  echo "  [ensure-dev-infra] 等待 Canal 进程（canal.deployer）…"
  while (( t < TIMEOUT )); do
    canal_process_running && {
      echo "  [ensure-dev-infra] Canal 进程已存在"
      return 0
    }
    (( t % 5 == 0 && t > 0 )) && echo "  [ensure-dev-infra] … 已等待 ${t}s"
    sleep 1
    t=$((t + 1))
  done
  echo "  [ensure-dev-infra] 错误: ${TIMEOUT}s 内 Canal 进程仍未出现" >&2
  return 1
}

ROCKETMQ_HOME="${ROCKETMQ_HOME:-$HOME/rocketmq}"
CANAL_HOME="${CANAL_HOME:-/usr/local/canal.deployer-1.1.8}"
ROCKETMQ_BROKER_HOST="${ROCKETMQ_BROKER_HOST:-127.0.0.1}"
typeset -i ROCKETMQ_BROKER_PORT=${ROCKETMQ_BROKER_PORT:-10911}

mq="${ROCKETMQ_NAMESRV:-127.0.0.1:9876}"
mq="${mq%%;*}"
mq="${mq#http://}"
mq="${mq#https://}"
mhost="${mq%%:*}"
mport="${mq##*:}"
[[ "$mport" == "$mhost" ]] && mport=9876

es="${ES_URI:-https://127.0.0.1:9200}"
read -r es_host es_port <<<"$(parse_http_host_port "$es" 9200)"

echo "==> ensure-dev-infra：按需启动 ES / RocketMQ / Canal（超时每项 ${TIMEOUT}s）"

# --- Elasticsearch ---
if ! tcp_open "$es_host" "$es_port"; then
  if [[ "$es_host" == "127.0.0.1" || "$es_host" == "localhost" ]]; then
    echo "  [ensure-dev-infra] 本机 Elasticsearch 未监听，尝试 brew services start …"
    if command -v brew >/dev/null 2>&1; then
      brew services start elastic/tap/elasticsearch-full 2>/dev/null || brew services start elasticsearch-full 2>/dev/null || true
    else
      echo "  [ensure-dev-infra] 未找到 brew，无法自动启动本机 ES；请手动启动或改 ES_URI" >&2
    fi
  else
    echo "  [ensure-dev-infra] 远程 ES (${es_host}:${es_port}) 不可连，不会在本机 brew 启动；请确保集群已就绪" >&2
  fi
  wait_tcp "Elasticsearch" "$es_host" "$es_port" || exit 1
else
  echo "  [ensure-dev-infra] Elasticsearch 已在监听 (${es_host}:${es_port})"
fi

# --- RocketMQ NameServer ---
if ! tcp_open "$mhost" "$mport"; then
  if [[ ! -d "$ROCKETMQ_HOME" ]]; then
    echo "  [ensure-dev-infra] RocketMQ 目录不存在: $ROCKETMQ_HOME（请安装或设置 ROCKETMQ_HOME）" >&2
    exit 1
  fi
  mkdir -p "$ROCKETMQ_HOME/logs"
  echo "  [ensure-dev-infra] 启动 RocketMQ NameServer …"
  nohup sh "$ROCKETMQ_HOME/bin/mqnamesrv" >>"$ROCKETMQ_HOME/logs/namesrv.log" 2>&1 &
  wait_tcp "RocketMQ NameServer" "$mhost" "$mport" || exit 1
else
  echo "  [ensure-dev-infra] RocketMQ NameServer 已在监听 (${mhost}:${mport})"
fi

# --- RocketMQ Broker ---
if ! tcp_open "$ROCKETMQ_BROKER_HOST" "$ROCKETMQ_BROKER_PORT"; then
  if [[ ! -d "$ROCKETMQ_HOME" ]]; then
    echo "  [ensure-dev-infra] 无法启动 Broker：ROCKETMQ_HOME 无效" >&2
    exit 1
  fi
  mkdir -p "$ROCKETMQ_HOME/logs"
  echo "  [ensure-dev-infra] 启动 RocketMQ Broker …"
  sleep 2
  nohup sh "$ROCKETMQ_HOME/bin/mqbroker" -n "${mhost}:${mport}" >>"$ROCKETMQ_HOME/logs/broker.log" 2>&1 &
  wait_tcp "RocketMQ Broker" "$ROCKETMQ_BROKER_HOST" "$ROCKETMQ_BROKER_PORT" || exit 1
else
  echo "  [ensure-dev-infra] RocketMQ Broker 已在监听 (${ROCKETMQ_BROKER_HOST}:${ROCKETMQ_BROKER_PORT})"
fi

# --- Canal ---
if canal_process_running; then
  echo "  [ensure-dev-infra] Canal 进程已在运行"
else
  if [[ ! -d "$CANAL_HOME" ]]; then
    echo "  [ensure-dev-infra] Canal 目录不存在: $CANAL_HOME（请安装或设置 CANAL_HOME）" >&2
    exit 1
  fi
  echo "  [ensure-dev-infra] 启动 Canal …"
  sh "$CANAL_HOME/bin/startup.sh"
  sleep 5
  canal_process_running || wait_canal_process || exit 1
fi

echo "==> ensure-dev-infra 完成"
exit 0

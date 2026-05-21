#!/bin/zsh
# 本地开发：若 NATS JetStream、Meilisearch 未就绪则尝试拉起。
#
#   SKIP_ENSURE_INFRA=1       跳过本脚本
#   NATS_URL                  默认 nats://127.0.0.1:4222
#   MEILI_HOST                默认 http://127.0.0.1:7700
#   MEILI_API_KEY             默认 go-migration-platform
#   MEILI_MASTER_KEY          未设置 MEILI_API_KEY 时作为 Meilisearch 密钥
#   ENSURE_INFRA_TIMEOUT      每项最长等待秒数，默认 120
#   ENSURE_INFRA_AUTO_INSTALL=1
#                             本机缺少 nats-server / meilisearch 时，尝试 brew install

[[ -n "$SKIP_ENSURE_INFRA" ]] && {
  echo "  [ensure-dev-infra] 已跳过（SKIP_ENSURE_INFRA）"
  exit 0
}

set -e

ROOT="${0:A:h}/.."
cd "${ROOT:A}"
mkdir -p .runlogs/nats-data .runlogs/meili-data

typeset -i TIMEOUT=${ENSURE_INFRA_TIMEOUT:-120}

tcp_open() {
  nc -z -w 2 "$1" "$2" >/dev/null 2>&1
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

is_local_host() {
  [[ "$1" == "127.0.0.1" || "$1" == "localhost" || "$1" == "::1" ]]
}

path_prepend_if_dir() {
  local dir="$1"
  [[ -d "$dir" ]] || return 0
  case ":$PATH:" in
    *":$dir:"*) ;;
    *) PATH="$dir:$PATH" ;;
  esac
}

ensure_command() {
  local command_name="$1" brew_formula="$2"
  if command -v "$command_name" >/dev/null 2>&1; then
    return 0
  fi
  if command -v brew >/dev/null 2>&1; then
    local brew_prefix=""
    local formula_prefix=""
    brew_prefix="$(brew --prefix 2>/dev/null || true)"
    formula_prefix="$(brew --prefix "$brew_formula" 2>/dev/null || true)"
    path_prepend_if_dir "${brew_prefix}/bin"
    path_prepend_if_dir "${brew_prefix}/sbin"
    path_prepend_if_dir "${formula_prefix}/bin"
    path_prepend_if_dir "${formula_prefix}/sbin"
    hash -r 2>/dev/null || true
    if command -v "$command_name" >/dev/null 2>&1; then
      return 0
    fi
  fi
  if [[ -n "$ENSURE_INFRA_AUTO_INSTALL" && -n "$brew_formula" && -x "$(command -v brew)" ]]; then
    echo "  [ensure-dev-infra] 未找到 $command_name，尝试 brew install $brew_formula …"
    brew install "$brew_formula" >/dev/null
    command -v "$command_name" >/dev/null 2>&1 && return 0
  fi
  return 1
}

nats_url="${NATS_URL:-nats://127.0.0.1:4222}"
read -r nats_host nats_port <<<"$(parse_nats_host_port "$nats_url")"
meili_host="${MEILI_HOST:-http://127.0.0.1:7700}"
meili_key="${MEILI_API_KEY:-${MEILI_MASTER_KEY:-go-migration-platform}}"
read -r meili_addr meili_port <<<"$(parse_http_host_port "$meili_host" 7700)"

echo "==> ensure-dev-infra：按需启动 NATS JetStream / Meilisearch（超时每项 ${TIMEOUT}s）"

if tcp_open "$nats_host" "$nats_port"; then
  echo "  [ensure-dev-infra] NATS 已在监听 (${nats_host}:${nats_port})"
elif is_local_host "$nats_host"; then
  if ! ensure_command nats-server nats-server; then
    echo "  [ensure-dev-infra] 未找到 nats-server。可执行 brew install nats-server，或设置 ENSURE_INFRA_AUTO_INSTALL=1 后重试。" >&2
    exit 1
  fi
  echo "  [ensure-dev-infra] 启动 NATS JetStream …"
  nohup nats-server -js -p "$nats_port" -sd .runlogs/nats-data > .runlogs/nats.log 2>&1 &
  wait_tcp "NATS JetStream" "$nats_host" "$nats_port" || exit 1
else
  wait_tcp "NATS JetStream" "$nats_host" "$nats_port" || exit 1
fi

if tcp_open "$meili_addr" "$meili_port"; then
  echo "  [ensure-dev-infra] Meilisearch 已在监听 (${meili_addr}:${meili_port})"
elif is_local_host "$meili_addr"; then
  if ! ensure_command meilisearch meilisearch; then
    echo "  [ensure-dev-infra] 未找到 meilisearch。可执行 brew install meilisearch，或设置 ENSURE_INFRA_AUTO_INSTALL=1 后重试。" >&2
    exit 1
  fi
  echo "  [ensure-dev-infra] 启动 Meilisearch …"
  MEILI_MASTER_KEY="$meili_key" nohup meilisearch \
    --http-addr "${meili_addr}:${meili_port}" \
    --master-key "$meili_key" \
    --db-path .runlogs/meili-data \
    > .runlogs/meilisearch.log 2>&1 &
  wait_tcp "Meilisearch" "$meili_addr" "$meili_port" || exit 1
else
  wait_tcp "Meilisearch" "$meili_addr" "$meili_port" || exit 1
fi

echo "==> ensure-dev-infra 完成"
exit 0

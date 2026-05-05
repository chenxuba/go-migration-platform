#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/deploy-backends.sh [options]

Options:
  --host <host>          SSH host, default: 43.240.15.181
  --user <user>          SSH user, default: root
  --port <port>          SSH port, default: 22
  --app-root <path>      Remote app root, default: /opt/go-migration-platform
  --services <list>      Skip prompts and deploy selected backend services.
                         Allowed names: all,iam,platform,education,8081,8082,8083
  --ports <list>         Alias of --services
  --tests                Run go test ./... before building
  --build-only           Build and package locally, do not upload or deploy
  -h, --help             Show this help

Environment:
  DEPLOY_SSH_PASSWORD    SSH password. Default is the embedded production server password.
  DEPLOY_GOOS            Backend build target OS, default: linux
  DEPLOY_GOARCH          Backend build target arch, default: amd64
  DEPLOY_SCP_LEGACY      1|0. Use legacy scp protocol by default.

Examples:
  ./scripts/deploy-backends.sh
  ./scripts/deploy-backends.sh --services 8082
  ./scripts/deploy-backends.sh --ports 8081,8083
  ./scripts/deploy-backends.sh --services all --tests
  ./scripts/deploy-backends.sh --services platform --build-only
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DEPLOY_HOST="${DEPLOY_HOST:-43.240.15.181}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_PORT="${DEPLOY_PORT:-22}"
DEPLOY_APP_ROOT="${DEPLOY_APP_ROOT:-/opt/go-migration-platform}"
DEPLOY_BACKENDS="${DEPLOY_BACKENDS:-}"
DEPLOY_BUILD_ONLY="${DEPLOY_BUILD_ONLY:-0}"
DEPLOY_RUN_TESTS="${DEPLOY_RUN_TESTS:-0}"
DEPLOY_GOOS="${DEPLOY_GOOS:-linux}"
DEPLOY_GOARCH="${DEPLOY_GOARCH:-amd64}"
DEPLOY_SCP_LEGACY="${DEPLOY_SCP_LEGACY:-1}"
DEPLOY_SSH_PASSWORD="${DEPLOY_SSH_PASSWORD:-i4SiMqAx4EQA}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host)
      DEPLOY_HOST="${2:?missing host}"
      shift 2
      ;;
    --user)
      DEPLOY_USER="${2:?missing user}"
      shift 2
      ;;
    --port)
      DEPLOY_PORT="${2:?missing port}"
      shift 2
      ;;
    --app-root)
      DEPLOY_APP_ROOT="${2:?missing app root}"
      shift 2
      ;;
    --services|--ports)
      DEPLOY_BACKENDS="${2:?missing backend service list}"
      shift 2
      ;;
    --tests)
      DEPLOY_RUN_TESTS=1
      shift
      ;;
    --build-only)
      DEPLOY_BUILD_ONLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

log() {
  printf '[deploy-backends] %s\n' "$*"
}

die() {
  printf '[deploy-backends] ERROR: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

append_backend() {
  local name="$1"
  if [[ -z "$DEPLOY_BACKENDS" ]]; then
    DEPLOY_BACKENDS="$name"
  else
    DEPLOY_BACKENDS="$DEPLOY_BACKENDS,$name"
  fi
}

ask_backend() {
  local name="$1"
  local port="$2"
  local label="$3"
  local answer

  while true; do
    printf '是否打包部署 %s（端口 %s）? [Y/n] ' "$label" "$port"
    read -r answer
    case "$answer" in
      ''|y|Y|yes|YES|Yes)
        append_backend "$name"
        break
        ;;
      n|N|no|NO|No)
        break
        ;;
      *)
        echo '请输入 y 或 n；直接回车表示 y'
        ;;
    esac
  done
}

choose_backends() {
  if [[ -n "$DEPLOY_BACKENDS" ]]; then
    return
  fi

  [[ -t 0 ]] || die "interactive backend selection requires a terminal; use --services all or --services 8082 instead"

  echo '请选择本次要打包部署的后端服务；直接一路回车就是 all：'
  ask_backend iam 8081 '8081 IAM/SSO 服务'
  ask_backend platform 8082 '8082 平台服务'
  ask_backend education 8083 '8083 教务服务'
}

normalize_backend_name() {
  case "$1" in
    all) echo "all" ;;
    iam|8081) echo "iam" ;;
    platform|8082) echo "platform" ;;
    education|8083) echo "education" ;;
    *) return 1 ;;
  esac
}

validate_backends() {
  local list="$DEPLOY_BACKENDS,"
  local raw
  local name
  local normalized=""

  while [[ -n "$list" ]]; do
    raw="${list%%,*}"
    list="${list#*,}"
    raw="${raw//[[:space:]]/}"
    [[ -z "$raw" ]] && continue

    if ! name="$(normalize_backend_name "$raw")"; then
      die "invalid backend service: $raw"
    fi

    if [[ "$name" == "all" ]]; then
      DEPLOY_BACKENDS="iam,platform,education"
      return
    fi

    case ",$normalized," in
      *",$name,"*) continue ;;
    esac
    if [[ -z "$normalized" ]]; then
      normalized="$name"
    else
      normalized="$normalized,$name"
    fi
  done

  DEPLOY_BACKENDS="$normalized"
}

should_include_backend() {
  case ",$DEPLOY_BACKENDS," in
    *",$1,"*) return 0 ;;
    *) return 1 ;;
  esac
}

backend_bin() {
  case "$1" in
    iam) echo "iam-service" ;;
    platform) echo "platform-service" ;;
    education) echo "education-service" ;;
  esac
}

backend_pkg() {
  case "$1" in
    iam) echo "./services/iam/cmd/api" ;;
    platform) echo "./services/platform/cmd/api" ;;
    education) echo "./services/education/cmd/api" ;;
  esac
}

backend_port() {
  case "$1" in
    iam) echo "8081" ;;
    platform) echo "8082" ;;
    education) echo "8083" ;;
  esac
}

run_go_backend() {
  local name="$1"
  local bin
  local pkg
  local port
  bin="$(backend_bin "$name")"
  pkg="$(backend_pkg "$name")"
  port="$(backend_port "$name")"

  log "build backend: $name port $port -> bin/$bin"
  (
    cd "$ROOT_DIR"
    CGO_ENABLED=0 GOOS="$DEPLOY_GOOS" GOARCH="$DEPLOY_GOARCH" go build -trimpath -ldflags "-s -w" -o "$PACKAGE_DIR/bin/$bin" "$pkg"
  )
}

write_remote_script() {
  cat > "$REMOTE_SCRIPT_LOCAL" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

archive="${1:?archive is required}"
release_id="${2:?release id is required}"
app_root="${3:?app root is required}"
selected_backends="${4:?selected backends are required}"

current_dir="$app_root/current"
env_file="/etc/go-migration-platform/app.env"
staging_dir="/tmp/go-migration-backends-$release_id"

log() {
  printf '[remote-backends] %s\n' "$*"
}

backend_bin() {
  case "$1" in
    iam) echo "iam-service" ;;
    platform) echo "platform-service" ;;
    education) echo "education-service" ;;
    *) return 1 ;;
  esac
}

backend_unit() {
  case "$1" in
    iam) echo "go-migration-iam.service" ;;
    platform) echo "go-migration-platform.service" ;;
    education) echo "go-migration-education.service" ;;
    *) return 1 ;;
  esac
}

backend_desc() {
  case "$1" in
    iam) echo "Go Migration IAM Service" ;;
    platform) echo "Go Migration Platform Service" ;;
    education) echo "Go Migration Education Service" ;;
    *) return 1 ;;
  esac
}

backend_port() {
  case "$1" in
    iam) echo "8081" ;;
    platform) echo "8082" ;;
    education) echo "8083" ;;
    *) return 1 ;;
  esac
}

install_unit_if_missing() {
  local name="$1"
  local unit
  local desc
  local bin
  local unit_path
  unit="$(backend_unit "$name")"
  desc="$(backend_desc "$name")"
  bin="$(backend_bin "$name")"
  unit_path="/etc/systemd/system/$unit"

  log "install systemd unit: $unit"
  cat > "$unit_path" <<UNIT
[Unit]
Description=$desc
After=network.target mysql95.service nats-server.service meilisearch.service
Wants=nats-server.service meilisearch.service

[Service]
Type=simple
EnvironmentFile=/etc/go-migration-platform/app.env
WorkingDirectory=$app_root/current
ExecStart=$app_root/current/bin/$bin
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
UNIT
}

[[ -f "$archive" ]] || { echo "archive not found: $archive" >&2; exit 1; }
[[ -f "$env_file" ]] || { echo "$env_file not found; keep server env config there before deploying" >&2; exit 1; }
[[ -e "$current_dir" ]] || { echo "current release not found: $current_dir" >&2; exit 1; }

rm -rf "$staging_dir"
mkdir -p "$staging_dir"

log "extract backend package: $release_id"
tar -xzf "$archive" -C "$staging_dir"
mkdir -p "$current_dir/bin"

list="$selected_backends,"
while [[ -n "$list" ]]; do
  name="${list%%,*}"
  list="${list#*,}"
  [[ -z "$name" ]] && continue

  bin="$(backend_bin "$name")" || { echo "invalid backend name: $name" >&2; exit 1; }
  src="$staging_dir/bin/$bin"
  dst="$current_dir/bin/$bin"
  tmp="$current_dir/bin/.$bin.deploying-$release_id"

  [[ -f "$src" ]] || { echo "backend package missing: bin/$bin" >&2; exit 1; }

  install_unit_if_missing "$name"

  log "replace $bin"
  rm -f "$tmp"
  mv "$src" "$tmp"
  chmod +x "$tmp"
  mv "$tmp" "$dst"
done

systemctl daemon-reload

list="$selected_backends,"
while [[ -n "$list" ]]; do
  name="${list%%,*}"
  list="${list#*,}"
  [[ -z "$name" ]] && continue

  unit="$(backend_unit "$name")"
  port="$(backend_port "$name")"
  log "restart $unit port $port"
  systemctl enable "$unit" >/dev/null
  systemctl restart "$unit"
done

list="$selected_backends,"
while [[ -n "$list" ]]; do
  name="${list%%,*}"
  list="${list#*,}"
  [[ -z "$name" ]] && continue

  unit="$(backend_unit "$name")"
  port="$(backend_port "$name")"
  systemctl is-active --quiet "$unit"
  log "$unit is active on port $port"
done

if command -v ss >/dev/null 2>&1; then
  ports=""
  list="$selected_backends,"
  while [[ -n "$list" ]]; do
    name="${list%%,*}"
    list="${list#*,}"
    [[ -z "$name" ]] && continue
    port="$(backend_port "$name")"
    if [[ -z "$ports" ]]; then
      ports=":$port"
    else
      ports="$ports|:$port"
    fi
  done
  if [[ -n "$ports" ]]; then
    ss -lntp | grep -E "$ports" || true
  fi
fi

rm -rf "$staging_dir"
rm -f "$archive"
rm -f "$0"
log "done: $release_id"
EOF
}

upload_with_expect() {
  local local_file="$1"
  local remote_file="$2"

  DEPLOY_EXPECT_PASSWORD="$DEPLOY_SSH_PASSWORD" \
  DEPLOY_EXPECT_HOST="$DEPLOY_HOST" \
  DEPLOY_EXPECT_USER="$DEPLOY_USER" \
  DEPLOY_EXPECT_PORT="$DEPLOY_PORT" \
  DEPLOY_EXPECT_LOCAL="$local_file" \
  DEPLOY_EXPECT_REMOTE="$remote_file" \
  DEPLOY_EXPECT_SCP_LEGACY="$DEPLOY_SCP_LEGACY" \
  expect <<'EOF'
set timeout -1
set cmd [list scp]
if {$env(DEPLOY_EXPECT_SCP_LEGACY) == "1"} {
  lappend cmd -O
}
lappend cmd -P $env(DEPLOY_EXPECT_PORT)
lappend cmd -o StrictHostKeyChecking=no
lappend cmd -o UserKnownHostsFile=/dev/null
lappend cmd $env(DEPLOY_EXPECT_LOCAL)
lappend cmd "$env(DEPLOY_EXPECT_USER)@$env(DEPLOY_EXPECT_HOST):$env(DEPLOY_EXPECT_REMOTE)"
eval spawn $cmd
expect {
  -re "yes/no" { send "yes\r"; exp_continue }
  -re "(P|p)assword:" { send "$env(DEPLOY_EXPECT_PASSWORD)\r"; exp_continue }
  eof
}
catch wait result
exit [lindex $result 3]
EOF
}

remote_run_with_expect() {
  DEPLOY_EXPECT_PASSWORD="$DEPLOY_SSH_PASSWORD" \
  DEPLOY_EXPECT_HOST="$DEPLOY_HOST" \
  DEPLOY_EXPECT_USER="$DEPLOY_USER" \
  DEPLOY_EXPECT_PORT="$DEPLOY_PORT" \
  DEPLOY_EXPECT_SCRIPT="$REMOTE_SCRIPT_REMOTE" \
  DEPLOY_EXPECT_ARCHIVE="$REMOTE_ARCHIVE" \
  DEPLOY_EXPECT_RELEASE="$RELEASE_ID" \
  DEPLOY_EXPECT_APP_ROOT="$DEPLOY_APP_ROOT" \
  DEPLOY_EXPECT_BACKENDS="$DEPLOY_BACKENDS" \
  expect <<'EOF'
set timeout -1
set cmd [list ssh]
lappend cmd -p $env(DEPLOY_EXPECT_PORT)
lappend cmd -o StrictHostKeyChecking=no
lappend cmd -o UserKnownHostsFile=/dev/null
lappend cmd "$env(DEPLOY_EXPECT_USER)@$env(DEPLOY_EXPECT_HOST)"
lappend cmd bash
lappend cmd $env(DEPLOY_EXPECT_SCRIPT)
lappend cmd $env(DEPLOY_EXPECT_ARCHIVE)
lappend cmd $env(DEPLOY_EXPECT_RELEASE)
lappend cmd $env(DEPLOY_EXPECT_APP_ROOT)
lappend cmd $env(DEPLOY_EXPECT_BACKENDS)
eval spawn $cmd
expect {
  -re "yes/no" { send "yes\r"; exp_continue }
  -re "(P|p)assword:" { send "$env(DEPLOY_EXPECT_PASSWORD)\r"; exp_continue }
  eof
}
catch wait result
exit [lindex $result 3]
EOF
}

upload_file() {
  local local_file="$1"
  local remote_file="$2"

  if [[ -n "$DEPLOY_SSH_PASSWORD" && "$USE_EXPECT" == "1" ]]; then
    upload_with_expect "$local_file" "$remote_file"
    return
  fi

  local scp_args=()
  if [[ "$DEPLOY_SCP_LEGACY" == "1" ]]; then
    scp_args+=("-O")
  fi
  scp_args+=("-P" "$DEPLOY_PORT" "-o" "StrictHostKeyChecking=no" "-o" "UserKnownHostsFile=/dev/null")

  if [[ -n "$DEPLOY_SSH_PASSWORD" && "$USE_SSHPASS" == "1" ]]; then
    sshpass -p "$DEPLOY_SSH_PASSWORD" scp "${scp_args[@]}" "$local_file" "$DEPLOY_USER@$DEPLOY_HOST:$remote_file"
  else
    scp "${scp_args[@]}" "$local_file" "$DEPLOY_USER@$DEPLOY_HOST:$remote_file"
  fi
}

remote_run() {
  if [[ -n "$DEPLOY_SSH_PASSWORD" && "$USE_EXPECT" == "1" ]]; then
    remote_run_with_expect
    return
  fi

  local ssh_args=("-p" "$DEPLOY_PORT" "-o" "StrictHostKeyChecking=no" "-o" "UserKnownHostsFile=/dev/null")
  if [[ -n "$DEPLOY_SSH_PASSWORD" && "$USE_SSHPASS" == "1" ]]; then
    sshpass -p "$DEPLOY_SSH_PASSWORD" ssh "${ssh_args[@]}" "$DEPLOY_USER@$DEPLOY_HOST" \
      bash "$REMOTE_SCRIPT_REMOTE" "$REMOTE_ARCHIVE" "$RELEASE_ID" "$DEPLOY_APP_ROOT" "$DEPLOY_BACKENDS"
  else
    ssh "${ssh_args[@]}" "$DEPLOY_USER@$DEPLOY_HOST" \
      bash "$REMOTE_SCRIPT_REMOTE" "$REMOTE_ARCHIVE" "$RELEASE_ID" "$DEPLOY_APP_ROOT" "$DEPLOY_BACKENDS"
  fi
}

choose_backends
validate_backends

if [[ -z "$DEPLOY_BACKENDS" ]]; then
  log "no backend selected; nothing to do"
  exit 0
fi

require_cmd tar
require_cmd go

if [[ "$DEPLOY_BUILD_ONLY" != "1" ]]; then
  require_cmd ssh
  require_cmd scp
fi

USE_SSHPASS=0
USE_EXPECT=0
if [[ "$DEPLOY_BUILD_ONLY" != "1" && -n "$DEPLOY_SSH_PASSWORD" ]]; then
  if command -v sshpass >/dev/null 2>&1; then
    USE_SSHPASS=1
  elif command -v expect >/dev/null 2>&1; then
    USE_EXPECT=1
  else
    die "DEPLOY_SSH_PASSWORD is set, but neither sshpass nor expect is installed"
  fi
fi

RELEASE_ID="$(date +%Y%m%d%H%M%S)"
WORK_DIR="$ROOT_DIR/.deploy/backends-$RELEASE_ID"
PACKAGE_DIR="$WORK_DIR/package"
ARCHIVE="$WORK_DIR/go-migration-backends-$RELEASE_ID.tgz"
REMOTE_ARCHIVE="/tmp/go-migration-backends-$RELEASE_ID.tgz"
REMOTE_SCRIPT_LOCAL="$WORK_DIR/remote-deploy-backends.sh"
REMOTE_SCRIPT_REMOTE="/tmp/go-migration-backends-remote-$RELEASE_ID.sh"

rm -rf "$WORK_DIR"
mkdir -p "$PACKAGE_DIR/bin"

log "selected backends: $DEPLOY_BACKENDS"
log "target: $DEPLOY_USER@$DEPLOY_HOST:$DEPLOY_PORT -> $DEPLOY_APP_ROOT"
log "release: $RELEASE_ID"

if [[ "$DEPLOY_RUN_TESTS" == "1" ]]; then
  log "run backend tests"
  (cd "$ROOT_DIR" && go test ./...)
fi

if should_include_backend iam; then
  run_go_backend iam
fi
if should_include_backend platform; then
  run_go_backend platform
fi
if should_include_backend education; then
  run_go_backend education
fi

write_remote_script

log "package backend release"
LC_ALL=C COPYFILE_DISABLE=1 tar -czf "$ARCHIVE" -C "$PACKAGE_DIR" .
chmod +x "$REMOTE_SCRIPT_LOCAL"

if [[ "$DEPLOY_BUILD_ONLY" == "1" ]]; then
  log "build-only archive: $ARCHIVE"
  log "completed"
  exit 0
fi

log "upload archive"
upload_file "$ARCHIVE" "$REMOTE_ARCHIVE"
log "upload remote script"
upload_file "$REMOTE_SCRIPT_LOCAL" "$REMOTE_SCRIPT_REMOTE"

log "deploy selected backends"
remote_run

log "completed"

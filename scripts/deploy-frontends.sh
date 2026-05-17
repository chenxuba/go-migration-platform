#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/deploy-frontends.sh [options]

Options:
  --host <host>          SSH host, default: 43.240.15.181
  --user <user>          SSH user, default: root
  --port <port>          SSH port, default: 22
  --app-root <path>      Remote app root, default: /opt/go-migration-platform
  --frontends <list>     Skip prompts and deploy selected frontends.
                         Allowed names: platform,institution,government,screen,training-games
  --skip-build           Reuse existing local dist directories
  --build-only           Build and package locally, do not upload or deploy
  -h, --help             Show this help

Environment:
  DEPLOY_SSH_PASSWORD    SSH password. Default is the embedded production server password.
  DEPLOY_INSTALL_DEPS    auto|1|0. Install frontend deps when node_modules is missing by default.
  DEPLOY_SCP_LEGACY      1|0. Use legacy scp protocol by default.

Examples:
  ./scripts/deploy-frontends.sh
  ./scripts/deploy-frontends.sh --frontends platform,institution
  ./scripts/deploy-frontends.sh --skip-build
  ./scripts/deploy-frontends.sh --build-only
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DEPLOY_HOST="${DEPLOY_HOST:-43.240.15.181}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_PORT="${DEPLOY_PORT:-22}"
DEPLOY_APP_ROOT="${DEPLOY_APP_ROOT:-/opt/go-migration-platform}"
DEPLOY_FRONTENDS="${DEPLOY_FRONTENDS:-}"
DEPLOY_BUILD_FRONTEND="${DEPLOY_BUILD_FRONTEND:-1}"
DEPLOY_BUILD_ONLY="${DEPLOY_BUILD_ONLY:-0}"
DEPLOY_INSTALL_DEPS="${DEPLOY_INSTALL_DEPS:-auto}"
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
    --frontends)
      DEPLOY_FRONTENDS="${2:?missing frontend list}"
      shift 2
      ;;
    --skip-build|--skip-frontend-build)
      DEPLOY_BUILD_FRONTEND=0
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
  printf '[deploy-frontends] %s\n' "$*"
}

die() {
  printf '[deploy-frontends] ERROR: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

append_frontend() {
  local name="$1"
  if [[ -z "$DEPLOY_FRONTENDS" ]]; then
    DEPLOY_FRONTENDS="$name"
  else
    DEPLOY_FRONTENDS="$DEPLOY_FRONTENDS,$name"
  fi
}

ask_frontend() {
  local name="$1"
  local label="$2"
  local dir="$3"
  local target="$4"
  local answer

  while true; do
    printf '是否打包部署 %s（%s -> web/%s）? [y/N] ' "$label" "$dir" "$target"
    read -r answer
    case "$answer" in
      y|Y|yes|YES|Yes)
        append_frontend "$name"
        break
        ;;
      n|N|no|NO|No|'')
        break
        ;;
      *)
        echo '请输入 y 或 n'
        ;;
    esac
  done
}

choose_frontends() {
  if [[ -n "$DEPLOY_FRONTENDS" ]]; then
    return
  fi

  [[ -t 0 ]] || die "interactive frontend selection requires a terminal; use --frontends platform,institution instead"

  echo '请选择本次要打包部署的前端项目：'
  ask_frontend platform '平台端' platform-admin platform
  ask_frontend institution '机构端' institution-admin institution
  ask_frontend government '监管端' government-admin government
  ask_frontend screen '数据大屏' government-screen screen
  ask_frontend training-games '训练小游戏' training-games training-games
}

validate_frontends() {
  local list="$DEPLOY_FRONTENDS,"
  local name
  local normalized=""

  while [[ -n "$list" ]]; do
    name="${list%%,*}"
    list="${list#*,}"
    name="${name//[[:space:]]/}"
    [[ -z "$name" ]] && continue
    case "$name" in
      platform|institution|government|screen|training-games)
        case ",$normalized," in
          *",$name,"*) continue ;;
        esac
        if [[ -z "$normalized" ]]; then
          normalized="$name"
        else
          normalized="$normalized,$name"
        fi
        ;;
      *) die "invalid frontend name: $name" ;;
    esac
  done

  DEPLOY_FRONTENDS="$normalized"
}

should_include_frontend() {
  case ",$DEPLOY_FRONTENDS," in
    *",$1,"*) return 0 ;;
    *) return 1 ;;
  esac
}

need_frontend_install() {
  local dir="$1"
  if [[ "$DEPLOY_INSTALL_DEPS" == "1" ]]; then
    return 0
  fi
  if [[ "$DEPLOY_INSTALL_DEPS" == "auto" && ! -d "$dir/node_modules" ]]; then
    return 0
  fi
  return 1
}

run_pnpm_frontend() {
  local dir="$1"
  local target="$2"
  local app_dir="$ROOT_DIR/$dir"
  local out_dir="$PACKAGE_DIR/web/$target"

  [[ -f "$app_dir/package.json" ]] || die "$dir/package.json not found"

  if [[ "$DEPLOY_BUILD_FRONTEND" == "1" ]]; then
    if need_frontend_install "$app_dir"; then
      log "install dependencies: $dir"
      (cd "$app_dir" && pnpm install --frozen-lockfile)
    fi
    log "build frontend: $dir -> web/$target"
    (cd "$app_dir" && pnpm build)
  else
    log "reuse frontend dist: $dir -> web/$target"
  fi

  [[ -d "$app_dir/dist" ]] || die "$dir/dist not found"
  mkdir -p "$out_dir"
  cp -a "$app_dir/dist/." "$out_dir/"
}

run_npm_frontend() {
  local dir="$1"
  local target="$2"
  local app_dir="$ROOT_DIR/$dir"
  local out_dir="$PACKAGE_DIR/web/$target"

  [[ -f "$app_dir/package.json" ]] || die "$dir/package.json not found"

  if [[ "$DEPLOY_BUILD_FRONTEND" == "1" ]]; then
    if need_frontend_install "$app_dir"; then
      log "install dependencies: $dir"
      (cd "$app_dir" && npm ci)
    fi
    log "build frontend: $dir -> web/$target"
    (cd "$app_dir" && npm run build)
  else
    log "reuse frontend dist: $dir -> web/$target"
  fi

  [[ -d "$app_dir/dist" ]] || die "$dir/dist not found"
  mkdir -p "$out_dir"
  cp -a "$app_dir/dist/." "$out_dir/"
}

write_remote_script() {
  cat > "$REMOTE_SCRIPT_LOCAL" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

archive="${1:?archive is required}"
release_id="${2:?release id is required}"
app_root="${3:?app root is required}"
selected_frontends="${4:?selected frontends are required}"

current_dir="$app_root/current"
staging_dir="/tmp/go-migration-frontends-$release_id"

log() {
  printf '[remote-frontends] %s\n' "$*"
}

[[ -f "$archive" ]] || { echo "archive not found: $archive" >&2; exit 1; }
[[ -e "$current_dir" ]] || { echo "current release not found: $current_dir" >&2; exit 1; }

rm -rf "$staging_dir"
mkdir -p "$staging_dir"

log "extract frontend package: $release_id"
tar -xzf "$archive" -C "$staging_dir"

mkdir -p "$current_dir/web"

list="$selected_frontends,"
while [[ -n "$list" ]]; do
  name="${list%%,*}"
  list="${list#*,}"
  [[ -z "$name" ]] && continue

  case "$name" in
    platform|institution|government|screen|training-games) ;;
    *) echo "invalid frontend name: $name" >&2; exit 1 ;;
  esac

  src="$staging_dir/web/$name"
  dst="$current_dir/web/$name"
  tmp="$current_dir/web/.$name.deploying-$release_id"

  [[ -d "$src" ]] || { echo "frontend package missing: web/$name" >&2; exit 1; }

  log "replace web/$name"
  rm -rf "$tmp"
  mv "$src" "$tmp"
  rm -rf "$dst"
  mv "$tmp" "$dst"
done

if command -v nginx >/dev/null 2>&1; then
  log "reload nginx"
  nginx -t
  systemctl reload nginx || nginx -s reload
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
  DEPLOY_EXPECT_FRONTENDS="$DEPLOY_FRONTENDS" \
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
lappend cmd $env(DEPLOY_EXPECT_FRONTENDS)
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
      bash "$REMOTE_SCRIPT_REMOTE" "$REMOTE_ARCHIVE" "$RELEASE_ID" "$DEPLOY_APP_ROOT" "$DEPLOY_FRONTENDS"
  else
    ssh "${ssh_args[@]}" "$DEPLOY_USER@$DEPLOY_HOST" \
      bash "$REMOTE_SCRIPT_REMOTE" "$REMOTE_ARCHIVE" "$RELEASE_ID" "$DEPLOY_APP_ROOT" "$DEPLOY_FRONTENDS"
  fi
}

choose_frontends
validate_frontends

if [[ -z "$DEPLOY_FRONTENDS" ]]; then
  log "no frontend selected; nothing to do"
  exit 0
fi

require_cmd tar

if should_include_frontend platform || should_include_frontend institution || should_include_frontend government; then
  if [[ "$DEPLOY_BUILD_FRONTEND" == "1" ]]; then
    require_cmd pnpm
  fi
fi
if should_include_frontend screen || should_include_frontend training-games; then
  if [[ "$DEPLOY_BUILD_FRONTEND" == "1" ]]; then
    require_cmd npm
  fi
fi

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
WORK_DIR="$ROOT_DIR/.deploy/frontends-$RELEASE_ID"
PACKAGE_DIR="$WORK_DIR/package"
ARCHIVE="$WORK_DIR/go-migration-frontends-$RELEASE_ID.tgz"
REMOTE_ARCHIVE="/tmp/go-migration-frontends-$RELEASE_ID.tgz"
REMOTE_SCRIPT_LOCAL="$WORK_DIR/remote-deploy-frontends.sh"
REMOTE_SCRIPT_REMOTE="/tmp/go-migration-frontends-remote-$RELEASE_ID.sh"

rm -rf "$WORK_DIR"
mkdir -p "$PACKAGE_DIR"

log "selected frontends: $DEPLOY_FRONTENDS"
log "target: $DEPLOY_USER@$DEPLOY_HOST:$DEPLOY_PORT -> $DEPLOY_APP_ROOT"
log "release: $RELEASE_ID"

if should_include_frontend platform; then
  run_pnpm_frontend platform-admin platform
fi
if should_include_frontend institution; then
  run_pnpm_frontend institution-admin institution
fi
if should_include_frontend government; then
  run_pnpm_frontend government-admin government
fi
if should_include_frontend screen; then
  run_npm_frontend government-screen screen
fi
if should_include_frontend training-games; then
  run_npm_frontend training-games training-games
fi

write_remote_script

log "package frontend release"
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

log "deploy selected frontends"
remote_run

log "completed"

#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/deploy-prod.sh [options]

Options:
  --host <host>          SSH host, default: 43.240.15.181
  --user <user>          SSH user, default: root
  --port <port>          SSH port, default: 22
  --app-root <path>      Remote app root, default: /opt/go-migration-platform
  --frontends <list>     Frontends to build and deploy, default: platform,institution,government,screen,training-games
                         Allowed names: platform,institution,government,screen,training-games
  --skip-frontend-build  Reuse existing local dist directories
  --tests                Run go test ./... before building
  --build-only           Build and package locally, do not upload or deploy
  -h, --help             Show this help

Environment:
  DEPLOY_SSH_PASSWORD    SSH password. Default is the embedded production server password.
  DEPLOY_GOOS            Backend build target OS, default: linux
  DEPLOY_GOARCH          Backend build target arch, default: amd64
  DEPLOY_KEEP_RELEASES   Number of remote releases to keep, default: 5
  DEPLOY_INSTALL_DEPS    auto|1|0. Install frontend deps when node_modules is missing by default.

Examples:
  ./scripts/deploy-prod.sh
  DEPLOY_SSH_PASSWORD='your-other-password' ./scripts/deploy-prod.sh
  ./scripts/deploy-prod.sh --frontends platform
  ./scripts/deploy-prod.sh --skip-frontend-build
  ./scripts/deploy-prod.sh --build-only
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DEPLOY_HOST="${DEPLOY_HOST:-43.240.15.181}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_PORT="${DEPLOY_PORT:-22}"
DEPLOY_APP_ROOT="${DEPLOY_APP_ROOT:-/opt/go-migration-platform}"
DEPLOY_FRONTENDS="${DEPLOY_FRONTENDS:-platform,institution,government,screen,training-games}"
DEPLOY_BUILD_FRONTEND="${DEPLOY_BUILD_FRONTEND:-1}"
DEPLOY_BUILD_ONLY="${DEPLOY_BUILD_ONLY:-0}"
DEPLOY_RUN_TESTS="${DEPLOY_RUN_TESTS:-0}"
DEPLOY_GOOS="${DEPLOY_GOOS:-linux}"
DEPLOY_GOARCH="${DEPLOY_GOARCH:-amd64}"
DEPLOY_KEEP_RELEASES="${DEPLOY_KEEP_RELEASES:-5}"
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
    --skip-frontend-build)
      DEPLOY_BUILD_FRONTEND=0
      shift
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
  printf '[deploy] %s\n' "$*"
}

die() {
  printf '[deploy] ERROR: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

should_include_frontend() {
  case ",$DEPLOY_FRONTENDS," in
    *",$1,"*) return 0 ;;
    *) return 1 ;;
  esac
}

validate_frontends() {
  local list="$DEPLOY_FRONTENDS,"
  local name

  while [[ -n "$list" ]]; do
    name="${list%%,*}"
    list="${list#*,}"
    [[ -z "$name" ]] && continue
    case "$name" in
      platform|institution|government|screen|training-games) ;;
      *) die "invalid frontend name: $name" ;;
    esac
  done
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

  if need_frontend_install "$app_dir"; then
    log "install dependencies: $dir"
    (cd "$app_dir" && pnpm install --frozen-lockfile)
  fi

  if [[ "$DEPLOY_BUILD_FRONTEND" == "1" ]]; then
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

  if need_frontend_install "$app_dir"; then
    log "install dependencies: $dir"
    (cd "$app_dir" && npm ci)
  fi

  if [[ "$DEPLOY_BUILD_FRONTEND" == "1" ]]; then
    log "build frontend: $dir -> web/$target"
    (cd "$app_dir" && npm run build)
  else
    log "reuse frontend dist: $dir -> web/$target"
  fi

  [[ -d "$app_dir/dist" ]] || die "$dir/dist not found"
  mkdir -p "$out_dir"
  cp -a "$app_dir/dist/." "$out_dir/"
}

build_backend() {
  local go_bin
  go_bin="$(command -v go)"

  if [[ "$DEPLOY_RUN_TESTS" == "1" ]]; then
    log "run backend tests"
    (cd "$ROOT_DIR" && "$go_bin" test ./...)
  fi

  log "build backend binaries for $DEPLOY_GOOS/$DEPLOY_GOARCH"
  mkdir -p "$PACKAGE_DIR/bin"
  (
    cd "$ROOT_DIR"
    CGO_ENABLED=0 GOOS="$DEPLOY_GOOS" GOARCH="$DEPLOY_GOARCH" "$go_bin" build -trimpath -ldflags "-s -w" -o "$PACKAGE_DIR/bin/iam-service" ./services/iam/cmd/api
    CGO_ENABLED=0 GOOS="$DEPLOY_GOOS" GOARCH="$DEPLOY_GOARCH" "$go_bin" build -trimpath -ldflags "-s -w" -o "$PACKAGE_DIR/bin/platform-service" ./services/platform/cmd/api
    CGO_ENABLED=0 GOOS="$DEPLOY_GOOS" GOARCH="$DEPLOY_GOARCH" "$go_bin" build -trimpath -ldflags "-s -w" -o "$PACKAGE_DIR/bin/education-service" ./services/education/cmd/api
  )
}

copy_configs() {
  mkdir -p "$PACKAGE_DIR/configs"
  if [[ -f "$ROOT_DIR/configs/tenants.example.json" ]]; then
    cp "$ROOT_DIR/configs/tenants.example.json" "$PACKAGE_DIR/configs/tenants.example.json"
  fi
  if [[ -f "$ROOT_DIR/configs/tenants.json" ]]; then
    cp "$ROOT_DIR/configs/tenants.json" "$PACKAGE_DIR/configs/tenants.json"
  fi

  local docs=()
  while IFS= read -r -d '' file; do
    docs+=("$file")
  done < <(find "$ROOT_DIR/docs" -maxdepth 1 -type f \( -name 'pep3*.json' -o -name 'erxin*.json' -o -name 'autismdev*.json' -o -name 'shuangxi*.json' \) -print0)
  if [[ "${#docs[@]}" -gt 0 ]]; then
    mkdir -p "$PACKAGE_DIR/docs"
    cp "${docs[@]}" "$PACKAGE_DIR/docs/"
  fi
}

write_remote_script() {
  cat > "$REMOTE_SCRIPT_LOCAL" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

archive="${1:?archive is required}"
release_id="${2:?release id is required}"
app_root="${3:?app root is required}"
keep_releases="${4:?keep releases is required}"

release_dir="$app_root/releases/$release_id"
current_dir="$app_root/current"
env_file="/etc/go-migration-platform/app.env"

log() {
  printf '[remote] %s\n' "$*"
}

install_unit_if_missing() {
  local unit="$1"
  local desc="$2"
  local bin="$3"
  local unit_path="/etc/systemd/system/$unit"

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

mkdir -p "$app_root/releases"
rm -rf "$release_dir"
mkdir -p "$release_dir"

log "extract release: $release_id"
tar -xzf "$archive" -C "$release_dir"

if [[ ! -f "$release_dir/configs/tenants.json" ]]; then
  if [[ -f "$current_dir/configs/tenants.json" ]]; then
    log "reuse existing configs/tenants.json"
    cp "$current_dir/configs/tenants.json" "$release_dir/configs/tenants.json"
  elif [[ -f "$release_dir/configs/tenants.example.json" ]]; then
    log "create configs/tenants.json from example"
    cp "$release_dir/configs/tenants.example.json" "$release_dir/configs/tenants.json"
  fi
fi

for name in platform institution government screen training-games; do
  if [[ ! -d "$release_dir/web/$name" && -d "$current_dir/web/$name" ]]; then
    log "reuse existing web/$name"
    mkdir -p "$release_dir/web"
    cp -a "$current_dir/web/$name" "$release_dir/web/$name"
  fi
done

install_unit_if_missing go-migration-iam.service "Go Migration IAM Service" iam-service
install_unit_if_missing go-migration-platform.service "Go Migration Platform Service" platform-service
install_unit_if_missing go-migration-education.service "Go Migration Education Service" education-service

ln -sfn "$release_dir" "$current_dir"

systemctl daemon-reload
systemctl enable go-migration-iam go-migration-platform go-migration-education >/dev/null

log "restart backend services"
systemctl restart go-migration-iam go-migration-platform go-migration-education

if command -v nginx >/dev/null 2>&1; then
  log "reload nginx"
  nginx -t
  systemctl reload nginx || nginx -s reload
fi

log "check service status"
systemctl is-active --quiet go-migration-iam
systemctl is-active --quiet go-migration-platform
systemctl is-active --quiet go-migration-education

if command -v ss >/dev/null 2>&1; then
  ss -lntp | grep -E ':8081|:8082|:8083' || true
fi

if [[ "$keep_releases" =~ ^[0-9]+$ && "$keep_releases" -gt 0 ]]; then
  find "$app_root/releases" -mindepth 1 -maxdepth 1 -type d | sort | head -n "-$keep_releases" | xargs -r rm -rf
fi

rm -f "$archive"
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
  DEPLOY_EXPECT_KEEP="$DEPLOY_KEEP_RELEASES" \
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
lappend cmd $env(DEPLOY_EXPECT_KEEP)
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
      bash "$REMOTE_SCRIPT_REMOTE" "$REMOTE_ARCHIVE" "$RELEASE_ID" "$DEPLOY_APP_ROOT" "$DEPLOY_KEEP_RELEASES"
  else
    ssh "${ssh_args[@]}" "$DEPLOY_USER@$DEPLOY_HOST" \
      bash "$REMOTE_SCRIPT_REMOTE" "$REMOTE_ARCHIVE" "$RELEASE_ID" "$DEPLOY_APP_ROOT" "$DEPLOY_KEEP_RELEASES"
  fi
}

require_cmd tar
require_cmd go
require_cmd ssh
require_cmd scp

validate_frontends

if should_include_frontend platform || should_include_frontend institution || should_include_frontend government; then
  require_cmd pnpm
fi
if should_include_frontend screen || should_include_frontend training-games; then
  require_cmd npm
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
WORK_DIR="$ROOT_DIR/.deploy/$RELEASE_ID"
PACKAGE_DIR="$WORK_DIR/package"
ARCHIVE="$WORK_DIR/go-migration-platform-$RELEASE_ID.tgz"
REMOTE_ARCHIVE="/tmp/go-migration-platform-$RELEASE_ID.tgz"
REMOTE_SCRIPT_LOCAL="$WORK_DIR/remote-deploy.sh"
REMOTE_SCRIPT_REMOTE="/tmp/go-migration-platform-remote-$RELEASE_ID.sh"

rm -rf "$WORK_DIR"
mkdir -p "$PACKAGE_DIR"

log "target: $DEPLOY_USER@$DEPLOY_HOST:$DEPLOY_PORT -> $DEPLOY_APP_ROOT"
log "release: $RELEASE_ID"

build_backend
copy_configs

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

log "package release"
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

log "deploy remote release"
remote_run

log "completed"

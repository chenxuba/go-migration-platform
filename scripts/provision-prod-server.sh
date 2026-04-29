#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/provision-prod-server.sh [options]

Purpose:
  Bootstrap a fresh production server end to end:
  - install nginx
  - install MySQL
  - install NATS JetStream
  - install Meilisearch
  - write /etc/go-migration-platform/app.env
  - write nginx site config
  - optionally import a SQL dump
  - optionally deploy the current project after provisioning

Recommended workflow:
  1. Copy scripts/provision-prod.env.example to scripts/provision-prod.env
  2. Fill in domain / SSH / DB / SSL values once
  3. Run:
       ./scripts/provision-prod-server.sh

Options:
  --config <path>          Provision config file. Default: scripts/provision-prod.env
  --host <host>            SSH host
  --user <user>            SSH user
  --port <port>            SSH port
  --app-root <path>        Remote app root. Default: /opt/go-migration-platform
  --domain <domain>        Primary domain, required
  --server-names <names>   nginx server_name value
  --ssl-mode <mode>        provided|self-signed|none
  --ssl-cert <path>        Local fullchain.pem path when ssl-mode=provided
  --ssl-key <path>         Local privkey.pem path when ssl-mode=provided
  --db-name <name>         Database name
  --db-user <user>         Database user
  --db-password <pass>     Database password
  --db-import-file <path>  Local .sql or .sql.gz dump to import
  --app-env-source <path>  Use this file as /etc/go-migration-platform/app.env
  --extra-env-file <path>  Append extra env lines when generating app.env
  --frontends <list>       Passed through to deploy-prod.sh
  --skip-deploy            Only provision server; do not deploy project
  -h, --help               Show this help

Notes:
  - If ssl-mode=provided, the cert and key are uploaded to:
      /etc/nginx/ssl/<domain>/
  - If ssl-mode=self-signed, the script creates a self-signed cert automatically.
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_CONFIG_FILE="$ROOT_DIR/scripts/provision-prod.env"
CONFIG_FILE="${PROVISION_CONFIG_FILE:-$DEFAULT_CONFIG_FILE}"

for ((i = 1; i <= $#; i++)); do
  if [[ "${!i}" == "--config" ]]; then
    next_index=$((i + 1))
    if [[ $next_index -gt $# ]]; then
      echo "[provision-prod] ERROR: missing config path after --config" >&2
      exit 1
    fi
    CONFIG_FILE="${!next_index}"
    break
  fi
done

if [[ -f "$CONFIG_FILE" ]]; then
  # shellcheck disable=SC1090
  source "$CONFIG_FILE"
fi

DEPLOY_HOST="${DEPLOY_HOST:-43.240.15.181}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_PORT="${DEPLOY_PORT:-22}"
DEPLOY_APP_ROOT="${DEPLOY_APP_ROOT:-/opt/go-migration-platform}"
DEPLOY_SSH_PASSWORD="${DEPLOY_SSH_PASSWORD:-i4SiMqAx4EQA}"
DEPLOY_SCP_LEGACY="${DEPLOY_SCP_LEGACY:-1}"

PROVISION_DOMAIN="${PROVISION_DOMAIN:-}"
PROVISION_SERVER_NAMES="${PROVISION_SERVER_NAMES:-}"
PROVISION_SSL_MODE="${PROVISION_SSL_MODE:-none}"
PROVISION_SSL_CERT="${PROVISION_SSL_CERT:-}"
PROVISION_SSL_KEY="${PROVISION_SSL_KEY:-}"

PROVISION_DB_NAME="${PROVISION_DB_NAME:-ybk_rebuild_edu}"
PROVISION_DB_USER="${PROVISION_DB_USER:-go_migration}"
PROVISION_DB_PASSWORD="${PROVISION_DB_PASSWORD:-}"
PROVISION_DB_IMPORT_FILE="${PROVISION_DB_IMPORT_FILE:-}"

PROVISION_APP_ENV="${PROVISION_APP_ENV:-prod}"
PROVISION_TOKEN_SECRET="${PROVISION_TOKEN_SECRET:-}"
PROVISION_TOKEN_COOKIE_NAME="${PROVISION_TOKEN_COOKIE_NAME:-ybcToken}"
PROVISION_NATS_VERSION="${PROVISION_NATS_VERSION:-2.10.24}"
PROVISION_MEILI_VERSION="${PROVISION_MEILI_VERSION:-1.12.3}"
PROVISION_MEILI_API_KEY="${PROVISION_MEILI_API_KEY:-go-migration-platform}"

PROVISION_APP_ENV_SOURCE="${PROVISION_APP_ENV_SOURCE:-}"
PROVISION_EXTRA_ENV_FILE="${PROVISION_EXTRA_ENV_FILE:-}"
PROVISION_FRONTENDS="${PROVISION_FRONTENDS:-platform,institution,government,screen}"
PROVISION_SKIP_DEPLOY="${PROVISION_SKIP_DEPLOY:-0}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --config)
      CONFIG_FILE="${2:?missing config path}"
      shift 2
      ;;
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
    --domain)
      PROVISION_DOMAIN="${2:?missing domain}"
      shift 2
      ;;
    --server-names)
      PROVISION_SERVER_NAMES="${2:?missing server names}"
      shift 2
      ;;
    --ssl-mode)
      PROVISION_SSL_MODE="${2:?missing ssl mode}"
      shift 2
      ;;
    --ssl-cert)
      PROVISION_SSL_CERT="${2:?missing ssl cert path}"
      shift 2
      ;;
    --ssl-key)
      PROVISION_SSL_KEY="${2:?missing ssl key path}"
      shift 2
      ;;
    --db-name)
      PROVISION_DB_NAME="${2:?missing db name}"
      shift 2
      ;;
    --db-user)
      PROVISION_DB_USER="${2:?missing db user}"
      shift 2
      ;;
    --db-password)
      PROVISION_DB_PASSWORD="${2:?missing db password}"
      shift 2
      ;;
    --db-import-file)
      PROVISION_DB_IMPORT_FILE="${2:?missing db import file}"
      shift 2
      ;;
    --app-env-source)
      PROVISION_APP_ENV_SOURCE="${2:?missing app env source}"
      shift 2
      ;;
    --extra-env-file)
      PROVISION_EXTRA_ENV_FILE="${2:?missing extra env file}"
      shift 2
      ;;
    --frontends)
      PROVISION_FRONTENDS="${2:?missing frontends list}"
      shift 2
      ;;
    --skip-deploy)
      PROVISION_SKIP_DEPLOY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "[provision-prod] ERROR: unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

log() {
  printf '[provision-prod] %s\n' "$*"
}

die() {
  printf '[provision-prod] ERROR: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

random_hex() {
  local bytes="$1"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$bytes"
    return
  fi
  od -An -N "$bytes" -tx1 /dev/urandom | tr -d ' \n'
}

validate_identifier() {
  [[ "$2" =~ ^[A-Za-z0-9_]+$ ]] || die "$1 must match [A-Za-z0-9_]+"
}

append_assignment() {
  local file="$1" key="$2" value="$3"
  printf '%s=%q\n' "$key" "$value" >> "$file"
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
  DEPLOY_EXPECT_STAGE_DIR="$REMOTE_STAGE_DIR" \
  expect <<'EOF'
set timeout -1
set cmd [list ssh]
lappend cmd -p $env(DEPLOY_EXPECT_PORT)
lappend cmd -o StrictHostKeyChecking=no
lappend cmd -o UserKnownHostsFile=/dev/null
lappend cmd "$env(DEPLOY_EXPECT_USER)@$env(DEPLOY_EXPECT_HOST)"
lappend cmd bash $env(DEPLOY_EXPECT_SCRIPT) $env(DEPLOY_EXPECT_STAGE_DIR)
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

if [[ "$PROVISION_SSL_MODE" != "provided" && "$PROVISION_SSL_MODE" != "self-signed" && "$PROVISION_SSL_MODE" != "none" ]]; then
  die "PROVISION_SSL_MODE must be provided, self-signed, or none"
fi

[[ -n "$PROVISION_DOMAIN" ]] || die "PROVISION_DOMAIN is required"
validate_identifier "PROVISION_DB_NAME" "$PROVISION_DB_NAME"
validate_identifier "PROVISION_DB_USER" "$PROVISION_DB_USER"

if [[ -z "$PROVISION_SERVER_NAMES" ]]; then
  PROVISION_SERVER_NAMES="_ $PROVISION_DOMAIN *.$PROVISION_DOMAIN"
fi

if [[ -z "$PROVISION_DB_PASSWORD" ]]; then
  PROVISION_DB_PASSWORD="$(random_hex 16)"
  GENERATED_DB_PASSWORD=1
else
  GENERATED_DB_PASSWORD=0
fi

if [[ -z "$PROVISION_TOKEN_SECRET" ]]; then
  PROVISION_TOKEN_SECRET="$(random_hex 24)"
  GENERATED_TOKEN_SECRET=1
else
  GENERATED_TOKEN_SECRET=0
fi

if [[ "$PROVISION_SSL_MODE" == "provided" ]]; then
  [[ -f "$PROVISION_SSL_CERT" ]] || die "ssl-mode=provided requires PROVISION_SSL_CERT"
  [[ -f "$PROVISION_SSL_KEY" ]] || die "ssl-mode=provided requires PROVISION_SSL_KEY"
fi

if [[ -n "$PROVISION_DB_IMPORT_FILE" ]]; then
  [[ -f "$PROVISION_DB_IMPORT_FILE" ]] || die "db import file not found: $PROVISION_DB_IMPORT_FILE"
fi

if [[ -n "$PROVISION_APP_ENV_SOURCE" ]]; then
  [[ -f "$PROVISION_APP_ENV_SOURCE" ]] || die "app env source not found: $PROVISION_APP_ENV_SOURCE"
fi

if [[ -n "$PROVISION_EXTRA_ENV_FILE" ]]; then
  [[ -f "$PROVISION_EXTRA_ENV_FILE" ]] || die "extra env file not found: $PROVISION_EXTRA_ENV_FILE"
fi

require_cmd tar
require_cmd ssh
require_cmd scp
require_cmd expect

TIMESTAMP="$(date +%Y%m%d%H%M%S)"
WORK_DIR="$ROOT_DIR/.deploy/provision-$TIMESTAMP"
REMOTE_STAGE_DIR="/tmp/go-migration-bootstrap-$TIMESTAMP"
REMOTE_SCRIPT_REMOTE="$REMOTE_STAGE_DIR/remote-bootstrap.sh"
APP_ENV_LOCAL="$WORK_DIR/app.env"
NGINX_LOCAL="$WORK_DIR/go-migration-platform.nginx.conf"
PROVISION_ENV_LOCAL="$WORK_DIR/provision.env"
REMOTE_SCRIPT_LOCAL="$WORK_DIR/remote-bootstrap.sh"

mkdir -p "$WORK_DIR"

write_app_env() {
  if [[ -n "$PROVISION_APP_ENV_SOURCE" ]]; then
    cp "$PROVISION_APP_ENV_SOURCE" "$APP_ENV_LOCAL"
  else
    cat > "$APP_ENV_LOCAL" <<EOF
APP_ENV=$PROVISION_APP_ENV
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=$PROVISION_DB_NAME
DB_USER=$PROVISION_DB_USER
DB_PASSWORD=$PROVISION_DB_PASSWORD
TOKEN_SECRET=$PROVISION_TOKEN_SECRET
TOKEN_COOKIE_NAME=$PROVISION_TOKEN_COOKIE_NAME
IAM_SERVICE_TENANT_CONFIG_PATH=$DEPLOY_APP_ROOT/current/configs/tenants.json
PLATFORM_SERVICE_TENANT_CONFIG_PATH=$DEPLOY_APP_ROOT/current/configs/tenants.json
EDUCATION_SERVICE_TENANT_CONFIG_PATH=$DEPLOY_APP_ROOT/current/configs/tenants.json
IAM_SERVICE_PORT=8081
PLATFORM_SERVICE_PORT=8082
EDUCATION_SERVICE_PORT=8083
NATS_URL=nats://127.0.0.1:4222
MEILI_HOST=http://127.0.0.1:7700
MEILI_API_KEY=$PROVISION_MEILI_API_KEY
EOF
    if [[ -n "$PROVISION_EXTRA_ENV_FILE" ]]; then
      printf '\n' >> "$APP_ENV_LOCAL"
      cat "$PROVISION_EXTRA_ENV_FILE" >> "$APP_ENV_LOCAL"
    fi
  fi
}

write_nginx_conf() {
  {
    cat <<EOF
upstream go_migration_iam {
    server 127.0.0.1:8081;
}

upstream go_migration_platform {
    server 127.0.0.1:8082;
}

upstream go_migration_education {
    server 127.0.0.1:8083;
}

server {
    listen 80 default_server;
    listen [::]:80 default_server;
EOF
    if [[ "$PROVISION_SSL_MODE" != "none" ]]; then
      cat <<EOF
    listen 443 ssl http2 default_server;
    listen [::]:443 ssl http2 default_server;
EOF
    fi
    cat <<EOF
    server_name $PROVISION_SERVER_NAMES;
EOF
    if [[ "$PROVISION_SSL_MODE" != "none" ]]; then
      cat <<EOF

    ssl_certificate /etc/nginx/ssl/$PROVISION_DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/$PROVISION_DOMAIN/privkey.pem;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_protocols TLSv1.2 TLSv1.3;
EOF
    fi
    cat <<EOF
    client_max_body_size 100m;

    location = / {
        return 302 /institution/;
    }

    location = /platform { return 301 /platform/; }
    location = /institution { return 301 /institution/; }
    location = /government { return 301 /government/; }
    location = /screen { return 301 /screen/; }

    location ~* ^/(platform|institution|government|screen)/assets/ {
        root $DEPLOY_APP_ROOT/current/web;
        expires 30d;
        add_header Cache-Control "public, max-age=2592000, immutable" always;
        add_header Access-Control-Allow-Origin "*" always;
        add_header Timing-Allow-Origin "*" always;
        try_files \$uri =404;
    }

    location ~* ^/(platform|institution|government|screen)/(loading\\.js|logo\\.svg)\$ {
        root $DEPLOY_APP_ROOT/current/web;
        expires 30d;
        add_header Cache-Control "public, max-age=2592000, immutable" always;
        try_files \$uri =404;
    }

    location /platform/ {
        alias $DEPLOY_APP_ROOT/current/web/platform/;
        add_header Cache-Control "no-cache" always;
        try_files \$uri \$uri/ /platform/index.html;
    }

    location /institution/ {
        alias $DEPLOY_APP_ROOT/current/web/institution/;
        add_header Cache-Control "no-cache" always;
        try_files \$uri \$uri/ /institution/index.html;
    }

    location /government/ {
        alias $DEPLOY_APP_ROOT/current/web/government/;
        add_header Cache-Control "no-cache" always;
        try_files \$uri \$uri/ /government/index.html;
    }

    location /screen/ {
        alias $DEPLOY_APP_ROOT/current/web/screen/;
        add_header Cache-Control "no-cache" always;
        try_files \$uri \$uri/ /screen/index.html;
    }

    location ^~ /platform-api/ {
        proxy_pass http://go_migration_platform/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location ^~ /sso/sso/ {
        proxy_pass http://go_migration_iam/sso/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location ^~ /sso/ {
        proxy_pass http://go_migration_iam;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location ^~ /api/v1/auth/ { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/users { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/manage-users/ { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/government-users { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/government-roles { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/login-logs { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/departs/ { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/menus/ { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/roles/ { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/tenants/current { proxy_pass http://go_migration_iam; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }

    location = /api/v1/public/login-theme {
        proxy_pass http://go_migration_platform;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        add_header Cache-Control "no-store, no-cache, must-revalidate, max-age=0" always;
        add_header Pragma "no-cache" always;
        add_header Expires "0" always;
        add_header Vary "Host, X-Tenant-Domain" always;
    }

    location ^~ /api/v1/tenant/ { proxy_pass http://go_migration_platform; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/public/ { proxy_pass http://go_migration_platform; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/platform/ { proxy_pass http://go_migration_platform; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /api/v1/qiniu/ { proxy_pass http://go_migration_platform; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }
    location ^~ /sys { proxy_pass http://go_migration_platform; proxy_set_header Host \$host; proxy_set_header X-Real-IP \$remote_addr; proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto \$scheme; }

    location ^~ /api/ {
        proxy_pass http://go_migration_education;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location / {
        return 404;
    }
}
EOF
  } > "$NGINX_LOCAL"
}

write_provision_env() {
  : > "$PROVISION_ENV_LOCAL"
  append_assignment "$PROVISION_ENV_LOCAL" APP_ROOT "$DEPLOY_APP_ROOT"
  append_assignment "$PROVISION_ENV_LOCAL" DOMAIN "$PROVISION_DOMAIN"
  append_assignment "$PROVISION_ENV_LOCAL" SSL_MODE "$PROVISION_SSL_MODE"
  append_assignment "$PROVISION_ENV_LOCAL" SSL_CERT_DIR "/etc/nginx/ssl/$PROVISION_DOMAIN"
  append_assignment "$PROVISION_ENV_LOCAL" DB_NAME "$PROVISION_DB_NAME"
  append_assignment "$PROVISION_ENV_LOCAL" DB_USER "$PROVISION_DB_USER"
  append_assignment "$PROVISION_ENV_LOCAL" DB_PASSWORD "$PROVISION_DB_PASSWORD"
  append_assignment "$PROVISION_ENV_LOCAL" DB_IMPORT_REMOTE "${PROVISION_DB_IMPORT_FILE:+$REMOTE_STAGE_DIR/$(basename "$PROVISION_DB_IMPORT_FILE")}"
  append_assignment "$PROVISION_ENV_LOCAL" NATS_VERSION "$PROVISION_NATS_VERSION"
  append_assignment "$PROVISION_ENV_LOCAL" MEILI_VERSION "$PROVISION_MEILI_VERSION"
  append_assignment "$PROVISION_ENV_LOCAL" MEILI_API_KEY "$PROVISION_MEILI_API_KEY"
}

write_remote_script() {
  cat > "$REMOTE_SCRIPT_LOCAL" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

stage_dir="${1:?stage dir is required}"
source "$stage_dir/provision.env"

log() {
  printf '[remote-bootstrap] %s\n' "$*"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing command: $1" >&2
    exit 1
  }
}

ensure_apt_packages() {
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y
  apt-get install -y ca-certificates curl gzip nginx openssl tar unzip
  apt-get install -y mysql-server
}

create_system_user() {
  local user="$1" home_dir="$2"
  if ! id -u "$user" >/dev/null 2>&1; then
    useradd --system --home "$home_dir" --shell /usr/sbin/nologin "$user"
  fi
}

install_nats_server() {
  local arch download_arch tmpdir version_out
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) download_arch="amd64" ;;
    aarch64|arm64) download_arch="arm64" ;;
    *) echo "unsupported arch for nats-server: $arch" >&2; exit 1 ;;
  esac

  version_out=""
  if command -v nats-server >/dev/null 2>&1; then
    version_out="$(nats-server -v 2>/dev/null || true)"
  fi
  if [[ "$version_out" == *"v$NATS_VERSION"* ]]; then
    return
  fi

  tmpdir="$(mktemp -d)"
  curl -fsSL -o "$tmpdir/nats-server.tgz" \
    "https://github.com/nats-io/nats-server/releases/download/v${NATS_VERSION}/nats-server-v${NATS_VERSION}-linux-${download_arch}.tar.gz"
  tar -xzf "$tmpdir/nats-server.tgz" -C "$tmpdir"
  install -m 0755 "$tmpdir/nats-server-v${NATS_VERSION}-linux-${download_arch}/nats-server" /usr/local/bin/nats-server
  rm -rf "$tmpdir"
}

install_meilisearch() {
  local arch download_arch version_out tmpdir url
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) download_arch="amd64" ;;
    aarch64|arm64) download_arch="aarch64" ;;
    *) echo "unsupported arch for meilisearch: $arch" >&2; exit 1 ;;
  esac

  version_out=""
  if command -v meilisearch >/dev/null 2>&1; then
    version_out="$(meilisearch --version 2>/dev/null || true)"
  fi
  if [[ "$version_out" == *"$MEILI_VERSION"* ]]; then
    return
  fi

  tmpdir="$(mktemp -d)"
  url="https://github.com/meilisearch/meilisearch/releases/download/v${MEILI_VERSION}/meilisearch-linux-${download_arch}"
  if ! curl -fsSL -o "$tmpdir/meilisearch" "$url"; then
    ( cd "$tmpdir" && curl -fsSL https://install.meilisearch.com | sh )
    mv "$tmpdir/meilisearch" "$tmpdir/meilisearch.bin"
  else
    mv "$tmpdir/meilisearch" "$tmpdir/meilisearch.bin"
  fi
  install -m 0755 "$tmpdir/meilisearch.bin" /usr/local/bin/meilisearch
  rm -rf "$tmpdir"
}

write_nats_unit() {
  cat > /etc/systemd/system/nats-server.service <<'UNIT'
[Unit]
Description=NATS Server with JetStream
After=network.target

[Service]
Type=simple
User=nats
Group=nats
ExecStart=/usr/local/bin/nats-server -js -p 4222 -sd /var/lib/nats/jetstream
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
UNIT
}

write_meilisearch_unit() {
  cat > /etc/systemd/system/meilisearch.service <<UNIT
[Unit]
Description=Meilisearch
After=network.target

[Service]
Type=simple
User=meilisearch
Group=meilisearch
WorkingDirectory=/var/lib/meilisearch
ExecStart=/usr/local/bin/meilisearch --http-addr 127.0.0.1:7700 --master-key ${MEILI_API_KEY} --db-path /var/lib/meilisearch/data
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
UNIT
}

install_runtime_dirs() {
  mkdir -p "$APP_ROOT/releases" "$APP_ROOT/backups" /etc/go-migration-platform
  mkdir -p /var/lib/nats/jetstream /var/lib/meilisearch/data
  create_system_user nats /var/lib/nats
  create_system_user meilisearch /var/lib/meilisearch
  chown -R nats:nats /var/lib/nats
  chown -R meilisearch:meilisearch /var/lib/meilisearch
}

configure_database() {
  local db_cli service_name password_sql
  db_cli="mysql"
  service_name="mysql"

  systemctl enable "$service_name" >/dev/null
  systemctl restart "$service_name"
  password_sql="$(printf "%s" "$DB_PASSWORD" | sed "s/'/''/g")"

  "$db_cli" <<SQL
CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$password_sql';
CREATE USER IF NOT EXISTS '$DB_USER'@'127.0.0.1' IDENTIFIED BY '$password_sql';
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'localhost';
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'127.0.0.1';
FLUSH PRIVILEGES;
SQL

  if [[ -n "$DB_IMPORT_REMOTE" && -f "$DB_IMPORT_REMOTE" ]]; then
    log "import database dump: $(basename "$DB_IMPORT_REMOTE")"
    if [[ "$DB_IMPORT_REMOTE" == *.gz ]]; then
      gzip -dc "$DB_IMPORT_REMOTE" | "$db_cli" "$DB_NAME"
    else
      "$db_cli" "$DB_NAME" < "$DB_IMPORT_REMOTE"
    fi
  fi
}

install_nginx_files() {
  install -m 0644 "$stage_dir/app.env" /etc/go-migration-platform/app.env

  mkdir -p /etc/nginx/conf.d /etc/nginx/sites-available /etc/nginx/sites-enabled
  cat > /etc/nginx/conf.d/go-migration-gzip-json.conf <<'CONF'
gzip_vary on;
gzip_proxied any;
gzip_comp_level 5;
gzip_min_length 1024;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
CONF
  cat > /etc/nginx/conf.d/go-migration-gzip-static.conf <<'CONF'
gzip_static on;
CONF

  install -m 0644 "$stage_dir/go-migration-platform.nginx.conf" /etc/nginx/sites-available/go-migration-platform
  ln -sfn /etc/nginx/sites-available/go-migration-platform /etc/nginx/sites-enabled/go-migration-platform
  rm -f /etc/nginx/sites-enabled/default

  case "$SSL_MODE" in
    provided)
      mkdir -p "$SSL_CERT_DIR"
      install -m 0600 "$stage_dir/fullchain.pem" "$SSL_CERT_DIR/fullchain.pem"
      install -m 0600 "$stage_dir/privkey.pem" "$SSL_CERT_DIR/privkey.pem"
      ;;
    self-signed)
      mkdir -p "$SSL_CERT_DIR"
      openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
        -keyout "$SSL_CERT_DIR/privkey.pem" \
        -out "$SSL_CERT_DIR/fullchain.pem" \
        -subj "/CN=$DOMAIN"
      chmod 600 "$SSL_CERT_DIR/privkey.pem" "$SSL_CERT_DIR/fullchain.pem"
      ;;
    none)
      ;;
    *)
      echo "unsupported SSL_MODE: $SSL_MODE" >&2
      exit 1
      ;;
  esac

  nginx -t
  systemctl enable nginx >/dev/null
  systemctl restart nginx
}

enable_runtime_services() {
  write_nats_unit
  write_meilisearch_unit
  systemctl daemon-reload
  systemctl enable nats-server meilisearch >/dev/null
  systemctl restart nats-server meilisearch
}

main() {
  require_cmd apt-get
  ensure_apt_packages
  install_runtime_dirs
  install_nats_server
  install_meilisearch
  enable_runtime_services
  configure_database
  install_nginx_files
  rm -rf "$stage_dir"
  log "server bootstrap complete"
}

main
EOF
  chmod +x "$REMOTE_SCRIPT_LOCAL"
}

write_app_env
write_nginx_conf
write_provision_env
write_remote_script

log "stage local files: $WORK_DIR"
log "target server: $DEPLOY_USER@$DEPLOY_HOST:$DEPLOY_PORT"
log "database engine: mysql"
log "ssl mode: $PROVISION_SSL_MODE"

expect <<EOF
set timeout -1
spawn ssh -p $DEPLOY_PORT -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $DEPLOY_USER@$DEPLOY_HOST "mkdir -p $REMOTE_STAGE_DIR"
expect {
  -re "yes/no" { send "yes\r"; exp_continue }
  -re "(P|p)assword:" { send "$DEPLOY_SSH_PASSWORD\r"; exp_continue }
  eof
}
catch wait result
exit [lindex \$result 3]
EOF

upload_with_expect "$APP_ENV_LOCAL" "$REMOTE_STAGE_DIR/app.env"
upload_with_expect "$NGINX_LOCAL" "$REMOTE_STAGE_DIR/go-migration-platform.nginx.conf"
upload_with_expect "$PROVISION_ENV_LOCAL" "$REMOTE_STAGE_DIR/provision.env"
upload_with_expect "$REMOTE_SCRIPT_LOCAL" "$REMOTE_SCRIPT_REMOTE"

if [[ "$PROVISION_SSL_MODE" == "provided" ]]; then
  upload_with_expect "$PROVISION_SSL_CERT" "$REMOTE_STAGE_DIR/fullchain.pem"
  upload_with_expect "$PROVISION_SSL_KEY" "$REMOTE_STAGE_DIR/privkey.pem"
fi

if [[ -n "$PROVISION_DB_IMPORT_FILE" ]]; then
  upload_with_expect "$PROVISION_DB_IMPORT_FILE" "$REMOTE_STAGE_DIR/$(basename "$PROVISION_DB_IMPORT_FILE")"
fi

log "bootstrap remote server"
remote_run_with_expect

if [[ "$PROVISION_SKIP_DEPLOY" != "1" ]]; then
  log "bootstrap finished; start full project deploy"
  DEPLOY_HOST="$DEPLOY_HOST" \
  DEPLOY_USER="$DEPLOY_USER" \
  DEPLOY_PORT="$DEPLOY_PORT" \
  DEPLOY_APP_ROOT="$DEPLOY_APP_ROOT" \
  DEPLOY_SSH_PASSWORD="$DEPLOY_SSH_PASSWORD" \
  "$ROOT_DIR/scripts/deploy-prod.sh" --frontends "$PROVISION_FRONTENDS"
fi

if [[ "${GENERATED_DB_PASSWORD:-0}" == "1" ]]; then
  log "generated DB password: $PROVISION_DB_PASSWORD"
fi
if [[ "${GENERATED_TOKEN_SECRET:-0}" == "1" ]]; then
  log "generated TOKEN_SECRET: $PROVISION_TOKEN_SECRET"
fi

log "done"

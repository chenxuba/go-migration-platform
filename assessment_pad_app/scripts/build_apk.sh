#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FIRST_ARG="${1:-local}"
PROD_DOMAIN="${PROD_DOMAIN:-irts-children.cn}"
PROD_SCHEME="${PROD_SCHEME:-https}"

if [[ "$FIRST_ARG" == "-h" || "$FIRST_ARG" == "--help" ]]; then
  cat <<'EOF'
Usage:
  scripts/build_apk.sh [local|prod] [release|debug|profile] [api-host]
  scripts/build_apk.sh [release|debug|profile] [api-host]

Examples:
  scripts/build_apk.sh local release 192.168.1.23
  scripts/build_apk.sh prod release
  scripts/build_apk.sh release 192.168.1.23
  API_HOST=192.168.1.23 scripts/build_apk.sh release
  LOGIN_API_BASE_URL=http://192.168.1.23:8081 EDUCATION_API_BASE_URL=http://192.168.1.23:8083 scripts/build_apk.sh release

Notes:
  Physical Android devices cannot use 127.0.0.1 to reach your Mac.
  Put the Pad and Mac on the same Wi-Fi, and use the Mac's LAN IP as api-host.
  prod defaults to https://irts-children.cn for both login and education APIs.
EOF
  exit 0
fi

if [[ "$FIRST_ARG" == "release" || "$FIRST_ARG" == "debug" || "$FIRST_ARG" == "profile" ]]; then
  BUILD_TARGET="${BUILD_TARGET:-local}"
  BUILD_MODE="$FIRST_ARG"
  API_HOST="${2:-${API_HOST:-}}"
else
  BUILD_TARGET="$FIRST_ARG"
  BUILD_MODE="${2:-release}"
  API_HOST="${3:-${API_HOST:-}}"
fi

if [[ "$BUILD_TARGET" != "local" && "$BUILD_TARGET" != "prod" && "$BUILD_TARGET" != "production" && "$BUILD_TARGET" != "online" ]]; then
  echo "Usage: scripts/build_apk.sh [local|prod] [release|debug|profile] [api-host]"
  exit 2
fi

if [[ "$BUILD_MODE" != "release" && "$BUILD_MODE" != "debug" && "$BUILD_MODE" != "profile" ]]; then
  echo "Usage: scripts/build_apk.sh [local|prod] [release|debug|profile] [api-host]"
  exit 2
fi

detect_lan_ip() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || true
    return
  fi
  if command -v hostname >/dev/null 2>&1; then
    hostname -I 2>/dev/null | awk '{print $1}'
  fi
}

if ! command -v flutter >/dev/null 2>&1; then
  echo "flutter command not found. Please add Flutter to PATH first."
  exit 1
fi

JDK_HOME="${JAVA_HOME:-}"

if [[ -z "$JDK_HOME" || ! -x "$JDK_HOME/bin/java" ]]; then
  if [[ "$(uname -s)" == "Darwin" && -x /usr/libexec/java_home ]]; then
    JDK_HOME="$(/usr/libexec/java_home -v 17 2>/dev/null || true)"
  fi
fi

if [[ -z "$JDK_HOME" || ! -x "$JDK_HOME/bin/java" ]]; then
  echo "JDK 17 was not found."
  echo "Install JDK 17, or set JAVA_HOME to your JDK 17 path, then run again."
  exit 1
fi

JAVA_VERSION_OUTPUT="$("$JDK_HOME/bin/java" -version 2>&1 | head -n 1)"
if [[ "$JAVA_VERSION_OUTPUT" != *\"17.* ]]; then
  echo "This project should be built with JDK 17 because its Gradle wrapper is 7.4.2."
  echo "Current JAVA_HOME: $JDK_HOME"
  echo "Current java version: $JAVA_VERSION_OUTPUT"
  echo "Set JAVA_HOME to JDK 17 and run again."
  exit 1
fi

export JAVA_HOME="$JDK_HOME"
export PATH="$JAVA_HOME/bin:$PATH"

cd "$PROJECT_DIR"

ensure_release_keystore() {
  local key_alias="${ANDROID_KEY_ALIAS:-assessment_assistant}"
  local store_file_rel="assessment-assistant-release.jks"
  local store_file="$PROJECT_DIR/android/app/$store_file_rel"
  local key_properties="$PROJECT_DIR/android/key.properties"
  local password

  if [[ -f "$key_properties" && -f "$store_file" ]]; then
    return
  fi

  password="${ANDROID_KEYSTORE_PASSWORD:-$(openssl rand -hex 16 2>/dev/null || uuidgen | tr -d '-')}"
  mkdir -p "$(dirname "$store_file")"
  "$JAVA_HOME/bin/keytool" -genkeypair -v \
    -keystore "$store_file" \
    -storetype JKS \
    -keyalg RSA \
    -keysize 2048 \
    -validity 10000 \
    -alias "$key_alias" \
    -storepass "$password" \
    -keypass "$password" \
    -dname "CN=Assessment Assistant, OU=YBK, O=YBK, L=Shenzhen, S=Guangdong, C=CN" >/dev/null

  cat > "$key_properties" <<EOF
storePassword=$password
keyPassword=$password
keyAlias=$key_alias
storeFile=$store_file_rel
EOF
}

if [[ "$BUILD_TARGET" == "prod" || "$BUILD_TARGET" == "production" || "$BUILD_TARGET" == "online" ]]; then
  BUILD_TARGET="prod"
  LOGIN_API_BASE_URL="${LOGIN_API_BASE_URL:-$PROD_SCHEME://$PROD_DOMAIN}"
  EDUCATION_API_BASE_URL="${EDUCATION_API_BASE_URL:-$PROD_SCHEME://$PROD_DOMAIN}"
  LOGIN_TENANT_DOMAIN="${LOGIN_TENANT_DOMAIN:-$PROD_DOMAIN}"
  LOGIN_QR_URL="${LOGIN_QR_URL:-$PROD_SCHEME://$PROD_DOMAIN/institution/}"
else
  BUILD_TARGET="local"
fi

if [[ "$BUILD_TARGET" == "local" && ( -z "${LOGIN_API_BASE_URL:-}" || -z "${EDUCATION_API_BASE_URL:-}" ) ]]; then
  if [[ -z "$API_HOST" ]]; then
    API_HOST="$(detect_lan_ip)"
  fi
  if [[ -z "$API_HOST" ]]; then
    echo "Could not detect a LAN IP."
    echo "Run with your Mac/backend IP, for example:"
    echo "  scripts/build_apk.sh release 192.168.1.23"
    exit 1
  fi
fi

if [[ "$BUILD_TARGET" == "local" ]]; then
  LOGIN_API_BASE_URL="${LOGIN_API_BASE_URL:-http://$API_HOST:8081}"
  EDUCATION_API_BASE_URL="${EDUCATION_API_BASE_URL:-http://$API_HOST:8083}"
fi

DEFAULT_LOGIN_USERNAME="${DEFAULT_LOGIN_USERNAME-17601241636}"
DEFAULT_LOGIN_PASSWORD="${DEFAULT_LOGIN_PASSWORD-123456}"

echo "Using JAVA_HOME=$JAVA_HOME"
echo "Using BUILD_TARGET=$BUILD_TARGET"
echo "Using LOGIN_API_BASE_URL=$LOGIN_API_BASE_URL"
echo "Using EDUCATION_API_BASE_URL=$EDUCATION_API_BASE_URL"
echo "Using DEFAULT_LOGIN_USERNAME=$DEFAULT_LOGIN_USERNAME"
flutter config --jdk-dir="$JAVA_HOME" >/dev/null

if [[ "$BUILD_TARGET" == "prod" && "$BUILD_MODE" == "release" ]]; then
  ensure_release_keystore
fi

DART_DEFINES=(
  "--dart-define=LOGIN_API_BASE_URL=$LOGIN_API_BASE_URL"
  "--dart-define=EDUCATION_API_BASE_URL=$EDUCATION_API_BASE_URL"
  "--dart-define=LOGIN_SOURCE=${LOGIN_SOURCE:-assessment-pad}"
  "--dart-define=DEFAULT_LOGIN_USERNAME=$DEFAULT_LOGIN_USERNAME"
  "--dart-define=DEFAULT_LOGIN_PASSWORD=$DEFAULT_LOGIN_PASSWORD"
)

if [[ -n "${LOGIN_TENANT_DOMAIN:-}" ]]; then
  DART_DEFINES+=("--dart-define=LOGIN_TENANT_DOMAIN=$LOGIN_TENANT_DOMAIN")
fi

if [[ -n "${LOGIN_QR_URL:-}" ]]; then
  DART_DEFINES+=("--dart-define=LOGIN_QR_URL=$LOGIN_QR_URL")
fi

flutter pub get
flutter build apk "--$BUILD_MODE" "${DART_DEFINES[@]}"

APK_PATH="$PROJECT_DIR/build/app/outputs/flutter-apk/app-$BUILD_MODE.apk"
if [[ -f "$APK_PATH" ]]; then
  if [[ "$BUILD_TARGET" == "prod" && "$BUILD_MODE" == "release" ]]; then
    FINAL_APK_PATH="$PROJECT_DIR/build/app/outputs/flutter-apk/评估助手.apk"
  else
    FINAL_APK_PATH="$PROJECT_DIR/build/app/outputs/flutter-apk/评估助手-$BUILD_TARGET-$BUILD_MODE.apk"
  fi
  cp -f "$APK_PATH" "$FINAL_APK_PATH"
  echo "APK built successfully:"
  echo "$APK_PATH"
  echo "$FINAL_APK_PATH"
else
  echo "Build finished, but APK was not found at expected path:"
  echo "$APK_PATH"
  exit 1
fi

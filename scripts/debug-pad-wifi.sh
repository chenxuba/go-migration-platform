#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/assessment_pad_app"

usage() {
  cat <<'USAGE'
用法：
  scripts/debug-pad-wifi.sh [PAD_IP[:PORT]] [额外 flutter run 参数]

这个脚本适用于“先用数据线连接 Pad，再切到无线调试”的流程。

执行步骤：
  1. adb kill-server
  2. adb start-server
  3. adb devices
  4. adb shell ip addr show wlan0
  5. adb tcpip 5555
  6. adb connect PAD_IP:5555
  7. adb devices
  8. flutter run -d PAD_IP:5555，并自动注入当前电脑 Wi-Fi IP 的后端地址

示例：
  scripts/debug-pad-wifi.sh
  scripts/debug-pad-wifi.sh 192.168.1.123
  scripts/debug-pad-wifi.sh 192.168.1.123:5555

可选环境变量：
  HOST_IP=192.168.1.203       手动指定当前电脑 IP
  PAD_IP=192.168.1.123        手动指定 Pad IP
  PAD_PORT=5555               手动指定 Pad 调试端口，默认 5555
  USB_SERIAL=xxxx             多台 USB 设备时手动指定数据线设备序列号
  LOGIN_PORT=8081             登录服务端口，默认 8081
  EDUCATION_PORT=8083         教育服务端口，默认 8083
USAGE
}

fail() {
  echo "error: $*" >&2
  exit 1
}

run_step() {
  echo
  echo "==> $*"
  "$@"
}

filter_flutter_warnings() {
  awk '
    BEGIN {
      spm_skip = 0
      kgp_skip = 0
    }
    /^The following plugins do not support Swift Package Manager for (ios|macos):$/ {
      spm_skip = 3
      next
    }
    spm_skip > 0 {
      spm_skip--
      next
    }
    /^WARNING: Your app uses the following plugins that apply Kotlin Gradle Plugin \(KGP\):/ {
      kgp_skip = 1
      next
    }
    kgp_skip == 1 {
      if ($0 ~ /for-plugin-authors$/) {
        kgp_skip = 0
      }
      next
    }
    { print }
  '
}

detect_host_ip() {
  local iface=""
  if command -v route >/dev/null 2>&1; then
    iface="$(route get default 2>/dev/null | awk '/interface:/{print $2; exit}')"
  fi

  for candidate in "$iface" en0 en1; do
    if [ -n "$candidate" ] && command -v ipconfig >/dev/null 2>&1; then
      ipconfig getifaddr "$candidate" 2>/dev/null && return 0
    fi
  done

  if command -v ifconfig >/dev/null 2>&1; then
    ifconfig | awk '/inet / && $2 !~ /^127\./ {print $2; exit}'
    return 0
  fi

  return 1
}

detect_usb_serial() {
  adb devices |
    awk '$2 == "device" && $1 !~ /:/ {print $1; exit}'
}

adb_for_usb() {
  if [ -n "${USB_SERIAL:-}" ]; then
    adb -s "$USB_SERIAL" "$@"
  else
    adb "$@"
  fi
}

extract_wlan0_ip() {
  awk '/inet / {
    split($2, parts, "/");
    print parts[1];
    exit;
  }'
}

normalize_device_id() {
  local raw="$1"
  local port="$2"
  if [[ "$raw" == *:* ]]; then
    echo "$raw"
  else
    echo "$raw:$port"
  fi
}

warn_if_port_closed() {
  local host="$1"
  local port="$2"
  local name="$3"
  if command -v nc >/dev/null 2>&1; then
    if ! nc -z -w 1 "$host" "$port" >/dev/null 2>&1; then
      echo "warning: $name 看起来不可访问：http://$host:$port" >&2
    fi
  fi
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi

PAD_ADDR="${1:-${PAD_IP:-}}"
if [ $# -gt 0 ] && [ -n "${1:-}" ]; then
  shift
fi

PAD_PORT="${PAD_PORT:-5555}"
HOST_IP="${HOST_IP:-$(detect_host_ip || true)}"
[ -n "$HOST_IP" ] || fail "无法自动识别当前电脑 IP，请用 HOST_IP=192.168.x.x 手动指定"

LOGIN_PORT="${LOGIN_PORT:-8081}"
EDUCATION_PORT="${EDUCATION_PORT:-8083}"
LOGIN_API_BASE_URL="http://$HOST_IP:$LOGIN_PORT"
EDUCATION_API_BASE_URL="http://$HOST_IP:$EDUCATION_PORT"

[ -f "$APP_DIR/pubspec.yaml" ] || fail "找不到 Flutter 项目：$APP_DIR/pubspec.yaml"
command -v adb >/dev/null 2>&1 || fail "找不到 adb 命令"
command -v flutter >/dev/null 2>&1 || fail "找不到 flutter 命令"

echo "电脑当前 IP:        $HOST_IP"
echo "LOGIN API:         $LOGIN_API_BASE_URL"
echo "EDUCATION API:     $EDUCATION_API_BASE_URL"

warn_if_port_closed "$HOST_IP" "$LOGIN_PORT" "LOGIN_API_BASE_URL"
warn_if_port_closed "$HOST_IP" "$EDUCATION_PORT" "EDUCATION_API_BASE_URL"

run_step adb kill-server
run_step adb start-server
run_step adb devices

if [ -z "${USB_SERIAL:-}" ]; then
  USB_SERIAL="$(detect_usb_serial || true)"
fi

if [ -z "${USB_SERIAL:-}" ]; then
  fail "没有检测到 USB 连接的 Pad。请先用数据线连接 Pad，并确认 Pad 已授权 USB 调试。"
fi

echo
echo "USB 设备:           $USB_SERIAL"

echo
echo "==> adb -s $USB_SERIAL shell ip addr show wlan0"
WLAN0_INFO="$(adb -s "$USB_SERIAL" shell ip addr show wlan0)"
echo "$WLAN0_INFO"

DETECTED_PAD_IP="$(printf '%s\n' "$WLAN0_INFO" | extract_wlan0_ip || true)"
if [ -z "$PAD_ADDR" ]; then
  PAD_ADDR="$DETECTED_PAD_IP"
fi

if [ -z "$PAD_ADDR" ]; then
  fail "没有从 wlan0 识别到 Pad IP，请手动执行：scripts/debug-pad-wifi.sh 192.168.1.xxx"
fi

DEVICE_ID="$(normalize_device_id "$PAD_ADDR" "$PAD_PORT")"

echo
echo "Pad 无线设备:       $DEVICE_ID"
if [ -n "$DETECTED_PAD_IP" ] && [[ "$DEVICE_ID" != "$DETECTED_PAD_IP:"* ]]; then
  echo "wlan0 识别 IP:      $DETECTED_PAD_IP"
fi

run_step adb -s "$USB_SERIAL" tcpip "$PAD_PORT"
sleep 1
run_step adb connect "$DEVICE_ID"
run_step adb devices

echo
echo "==> flutter run -d $DEVICE_ID"
cd "$APP_DIR"
set +e
flutter run -d "$DEVICE_ID" \
  --android-skip-build-dependency-validation \
  --dart-define=LOGIN_API_BASE_URL="$LOGIN_API_BASE_URL" \
  --dart-define=EDUCATION_API_BASE_URL="$EDUCATION_API_BASE_URL" \
  "$@" 2>&1 | filter_flutter_warnings
status=${PIPESTATUS[0]}
set -e
exit "$status"

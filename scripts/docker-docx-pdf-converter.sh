#!/bin/zsh
set -euo pipefail

if (( $# != 2 )); then
  echo "usage: $0 <input.docx> <output.pdf>" >&2
  exit 1
fi

input_path="$1"
output_path="$2"

if [[ ! -f "$input_path" ]]; then
  echo "input docx not found: $input_path" >&2
  exit 1
fi

input_dir="${input_path:A:h}"
input_file="${input_path:t}"
output_dir="${output_path:A:h}"
output_file="${output_path:t}"

mkdir -p "$output_dir"

script_dir="${0:A:h}"
docker_context="${script_dir}/docker/docx-pdf-converter"
image_tag="${DOCX_PDF_DOCKER_IMAGE:-go-migration-platform/docx-pdf-converter:local}"

ensure_docker() {
  command -v docker >/dev/null 2>&1 || {
    echo "docker command not found" >&2
    exit 1
  }
  docker info >/dev/null 2>&1 || {
    echo "docker daemon is not available" >&2
    exit 1
  }
}

ensure_image() {
  if docker image inspect "$image_tag" >/dev/null 2>&1; then
    return 0
  fi
  echo "[docker-docx-pdf] building image ${image_tag} ..."
  docker build -t "$image_tag" "$docker_context"
}

run_same_dir() {
  docker run --rm \
    -e HOME=/tmp \
    -e TMPDIR=/tmp \
    -v "${input_dir}:/work" \
    -w /work \
    "$image_tag" \
    --headless \
    --nologo \
    --nodefault \
    --nofirststartwizard \
    --nolockcheck \
    "-env:UserInstallation=file:///tmp/libreoffice-profile" \
    --convert-to pdf:writer_pdf_Export \
    --outdir /work \
    "/work/${input_file}"
}

run_split_dir() {
  docker run --rm \
    -e HOME=/tmp \
    -e TMPDIR=/tmp \
    -v "${input_dir}:/input:ro" \
    -v "${output_dir}:/output" \
    "$image_tag" \
    --headless \
    --nologo \
    --nodefault \
    --nofirststartwizard \
    --nolockcheck \
    "-env:UserInstallation=file:///tmp/libreoffice-profile" \
    --convert-to pdf:writer_pdf_Export \
    --outdir /output \
    "/input/${input_file}"
}

ensure_docker
ensure_image

if [[ "$input_dir" == "$output_dir" ]]; then
  run_same_dir
else
  run_split_dir
fi

if [[ ! -f "$output_path" ]]; then
  default_output="${output_dir}/${input_file:r}.pdf"
  if [[ "$default_output" != "$output_path" && -f "$default_output" ]]; then
    mv "$default_output" "$output_path"
  fi
fi

if [[ ! -f "$output_path" ]]; then
  echo "conversion finished but output pdf is missing: $output_path" >&2
  exit 1
fi

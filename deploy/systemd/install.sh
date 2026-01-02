#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Installs systemd units for google-flights workers (no Docker).

Usage:
  sudo ./deploy/systemd/install.sh --single
  sudo ./deploy/systemd/install.sh --instances 1 2 3

Notes:
  - Expects a built binary at /opt/google-flights/google-flights-api
  - Creates user/group "flights" and /etc/google-flights/worker.env (if missing)
EOF
}

if [[ ${1:-} == "-h" || ${1:-} == "--help" ]]; then
  usage
  exit 0
fi

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  echo "error: run as root (use sudo)" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

app_user="flights"
app_group="flights"
app_dir="/opt/google-flights"
env_dir="/etc/google-flights"
env_file="${env_dir}/worker.env"

install -d -m 0755 "${app_dir}" "${env_dir}"

if ! id -u "${app_user}" >/dev/null 2>&1; then
  useradd --system --home "${app_dir}" --shell /usr/sbin/nologin "${app_user}"
fi

if [[ ! -f "${env_file}" ]]; then
  install -m 0640 "${repo_root}/deploy/systemd/worker.env.example" "${env_file}"
  chown root:"${app_group}" "${env_file}" || true
  echo "created ${env_file} (edit required values before starting)"
fi

if [[ ${1:-} == "--single" ]]; then
  install -m 0644 "${repo_root}/deploy/systemd/google-flights-worker.service" /etc/systemd/system/google-flights-worker.service
  systemctl daemon-reload
  systemctl enable --now google-flights-worker.service
  systemctl status --no-pager google-flights-worker.service || true
  exit 0
fi

if [[ ${1:-} == "--instances" ]]; then
  shift
  if [[ $# -lt 1 ]]; then
    echo "error: provide one or more instance numbers (e.g. --instances 1 2 3)" >&2
    exit 1
  fi
  install -m 0644 "${repo_root}/deploy/systemd/google-flights-worker@.service" /etc/systemd/system/google-flights-worker@.service
  systemctl daemon-reload
  for instance in "$@"; do
    systemctl enable --now "google-flights-worker@${instance}.service"
  done
  systemctl status --no-pager "google-flights-worker@${1}.service" || true
  exit 0
fi

echo "error: expected --single or --instances" >&2
usage >&2
exit 1


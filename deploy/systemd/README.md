# systemd worker deployment

This directory contains `systemd` unit templates for running workers directly on a VPS (no Docker).

Files:
- `deploy/systemd/google-flights-worker.service`: single worker service
- `deploy/systemd/google-flights-worker@.service`: multi-instance template (`@1`, `@2`, ...)
- `deploy/systemd/worker.env.example`: environment file template (copy to `/etc/google-flights/worker.env`)

Minimal install steps (on the VPS):
```bash
sudo useradd --system --home /opt/google-flights --shell /usr/sbin/nologin flights || true
sudo mkdir -p /opt/google-flights /etc/google-flights

# Copy the built binary to: /opt/google-flights/google-flights-api
# Copy env file to: /etc/google-flights/worker.env
# Copy service file(s) to: /etc/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable --now google-flights-worker
```

To run multiple workers on one host:
```bash
sudo systemctl enable --now google-flights-worker@1 google-flights-worker@2
```

Optional helper (run from repo checkout on the VPS):
```bash
chmod +x deploy/systemd/install.sh
sudo ./deploy/systemd/install.sh --instances 1 2
```

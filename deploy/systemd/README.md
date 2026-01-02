# systemd worker deployment

This directory contains `systemd` unit templates for running workers directly on a VPS (no Docker).

Files:
- `deploy/systemd/google-flights-worker.service`: single worker service
- `deploy/systemd/google-flights-worker@.service`: multi-instance template (`@1`, `@2`, ...)
- `deploy/systemd/google-flights-worker-docker@.service`: multi-instance template using Docker images (no binary copy)
- `deploy/systemd/worker.env.example`: environment file template (copy to `/etc/google-flights/worker.env`)
- `deploy/systemd/worker.docker.env.example`: environment file template for Docker-based worker (copy to `/etc/google-flights/worker.docker.env`)

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

## Docker-based worker (no scripts / no binary copy)

If you prefer to avoid copying a Go binary to the worker VM, run the worker as a container and let `systemd` keep it alive.

Recommendation:
- Use the **built image** directly rather than extracting a binary from it.
- If you want a native binary on the VM, use the non-Docker units in this folder (`google-flights-worker@.service`).

1) Copy env file template and fill it in:
```bash
sudo mkdir -p /etc/google-flights
sudo cp deploy/systemd/worker.docker.env.example /etc/google-flights/worker.docker.env
sudo nano /etc/google-flights/worker.docker.env
```

2) Install the unit:
```bash
sudo cp deploy/systemd/google-flights-worker-docker@.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now google-flights-worker-docker@1
```

3) Logs:
```bash
journalctl -u google-flights-worker-docker@1 -f
```

## Komodo rule (do not violate)

Do **not** deploy, stop, start, rebuild, or wipe any **Komodo-managed** stack using Docker CLI commands (including `docker context`, `docker compose`, `docker rm`, `docker stop/start`, `docker system prune`, `down -v`, etc.).

Reason: Docker-side actions bypass/undermine Komodo’s control plane and can break Komodo deployments and desired-state management.

For Komodo environments, use:
- Komodo UI (or Komodo-native CLI) to deploy/stop/reset, and/or
- repo changes (compose/config) so Komodo deploys cleanly.

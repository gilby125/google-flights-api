# Repository Guidelines

## Project Structure & Module Organization
- `flights/` is the reusable Go client; keep it dependency-light.
- `api/`, `main/`, and `pkg/` expose HTTP handlers, CLI entry points, and shared utilities, while `worker/` and `queue/` coordinate background jobs.
- UI assets: `web/`, `static/`, `templates/`; tests: `test/`; config and tooling: `config/`, `db/`, `scripts/`; infra manifests at the root and in `kubernetes/`.

## Build, Test, and Development Commands
- `go run ./main` starts the API server (see `--help` for flags).
- `go test ./...` or `./run_tests.sh` runs all Go unit and integration tests.
- `npm install` (once) and `npm test` execute the Playwright suite; `npm run test:report` opens the report viewer.
- `docker-compose up -d --build` spins up the full stack (API, Postgres, Neo4j, Redis, Traefik) for end-to-end checks.

## Coding Style & Naming Conventions
- Run `gofmt` or `goimports` before committing; stick to idiomatic Go names and keep receivers concise.
- Align new folders with existing lower_snake_case conventions.
- Playwright specs should follow Prettier defaults (2 spaces, single quotes) and use `<feature>.spec.ts` filenames.
- Reference configuration via environment variables defined in `config/config.go`; never commit secrets.

## Testing Guidelines
- Co-locate Go tests as `<name>_test.go` using the standard `testing` package.
- Seed integration data through `db/seed.go` or helper fixtures so CI stays deterministic.
- Keep Playwright scenarios stable by resetting state within each test and avoid parallel mode unless selectors are isolated.

## Commit & Pull Request Guidelines
- Match the repo history: imperative subjects (`Fix retry handling`, `Add bulk search support`) with optional scopes like `fix:`.
- Limit each commit to one logical change and document migrations, new env vars, or API tweaks in the body.
- Pull requests should include context, test evidence (`go test`, `npm test`, screenshots for UI), and links to issues or tickets.
- Highlight breaking changes, new ports, or credential steps so reviewers can plan deployment updates.

## Security & Configuration Tips
- Keep credentials in `.env` (git-ignored) or a secret manager and align defaults with `config/config.go`.
- Rotate `DB_PASSWORD`, `NEO4J_PASSWORD`, and `REDIS_PASSWORD` when sharing environments, and avoid committing artifacts from `scripts/generate-tls-certs.sh`.
- Recheck Traefik and Kubernetes settings before exposing routes to ensure HTTPS, auth, and rate limits stay enforced.

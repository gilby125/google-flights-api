# Repository Guidelines

## Project Structure & Module Organization
Core Go logic lives in purpose-built modules, while assets and infrastructure files stay isolated for easy discovery.
- `flights/` provides the reusable Go client—keep it dependency-light and reusable.
- `api/`, `main/`, and `pkg/` deliver HTTP handlers, CLI entry points, and shared utilities; `worker/` and `queue/` orchestrate background jobs.
- Frontend assets live in `web/`, `static/`, and `templates/`; automated tests in `test/`.
- Configuration, data fixtures, and tooling are under `config/`, `db/`, and `scripts/`; deployment manifests sit at the root and in `kubernetes/`.

## Build, Test, and Development Commands
- `go run ./main` boots the API server (`--help` shows flags).
- `go test ./...` or `./run_tests.sh` exercises all Go unit and integration suites.
- `npm install` (once) then `npm test` runs the Playwright browser checks; `npm run test:report` opens the report UI.
- `docker-compose up -d --build` starts the full stack (API, Postgres, Neo4j, Redis, Traefik) for end-to-end validation.

## Coding Style & Naming Conventions
- Run `gofmt` or `goimports` on Go sources and prefer concise receiver names.
- Name new folders using lower_snake_case to align with current layout.
- Playwright specs should follow Prettier defaults (2 spaces, single quotes) and use filenames like `search_flow.spec.ts`.
- Read configuration keys from `config/config.go`; never hardcode secrets or credentials.

## Testing Guidelines
- Co-locate Go tests as `<name>_test.go` using the standard `testing` package.
- Seed integration data through `db/seed.go` or shared fixtures so CI remains stable.
- Reset state inside each Playwright test and avoid parallel mode unless selectors are isolated.

## Commit & Pull Request Guidelines
Document changes in a way reviewers can trust quickly.
- Write imperative commit subjects (`Add fare caching`, `Fix retry handling`) with optional scopes like `fix:`.
- Keep commits focused on one logical change and note migrations, env vars, or API tweaks in the body.
- Pull requests should describe the change, link related issues, and attach evidence (`go test`, `npm test`, screenshots for UI) before requesting review.
- Flag breaking changes, new ports, or credential updates so release prep stays smooth.

## Security & Configuration Tips
- Store secrets in `.env` or a secret manager, aligning defaults with `config/config.go`.
- Rotate `DB_PASSWORD`, `NEO4J_PASSWORD`, and `REDIS_PASSWORD` when sharing environments and keep artifacts from `scripts/generate-tls-certs.sh` out of Git.
- Review Traefik and Kubernetes settings before exposing routes to confirm HTTPS, auth, and rate limits stay enforced.

## Notes for AI Assistants
- You only see the files in this repo and can run local build/test commands, but you cannot see the user’s actual browser, Docker daemon, or other host processes.
- When debugging runtime issues (for example, network errors from the browser), assume the code here is correct unless logs or config suggest otherwise, and give the user explicit commands to run (`PORT=8081 go run ./main`, `docker compose up -d`, `curl` health checks).
- Treat these guidelines as context, not hard limits; you should still freely inspect the repo and use typical Go and Node tooling within the workspace.

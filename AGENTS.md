# Repository Guidelines for AI Agents

## Build & Test Commands
- `go run ./main` - Start API server (use `--help` for flags)
- `go test ./... -v` - Run all Go tests (or `./run_tests.sh`)
- `go test -v ./path/to/file_test.go` - Run single test file
- `go test ./... -coverprofile=coverage.out` - Generate coverage report
- `npm test` - Run Playwright E2E tests (requires `npm install`)
- `npm run test:headed` - Run Playwright with visible browser
- `docker compose up -d --build` - Start full stack for integration testing

## Code Style & Conventions
- Run `gofmt` or `goimports` on all Go sources before committing
- Use standard Go error handling: `if err != nil { return fmt.Errorf("context: %w", err) }`
- Name directories with `lower_snake_case`, files with `snake_case.go`
- Co-locate tests as `<name>_test.go` using standard `testing` package
- Use `testify` for assertions in tests, follow mock patterns in `test/mocks/`
- Import groups: stdlib, third-party, local (separated by blank lines)
- Read config from `config/config.go`; never hardcode secrets

## Architecture Notes
- `flights/` is the standalone client library (keep dependency-light)
- HTTP handlers in `api/` use Gin framework, routes in `routes.go`
- Background jobs: Redis queue (`queue/`) → workers (`worker/`) → databases
- Dual storage: PostgreSQL for results, Neo4j for route graphs
- All endpoints documented in `docs/API_CONTRACT.md`

## Testing Strategy
- Unit tests in `test/unit/`, integration in `test/integration/`, E2E in `test/e2e/`
- Use `skipUnlessIntegration(t)` for tests requiring external services
- Reset Playwright state between tests, avoid parallel mode unless isolated

## Operational Safety
- Do not deploy, restart, stop, or modify live services via CLI. Only deploy through the approved UI when explicitly instructed by the user.

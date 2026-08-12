# Repository Guidelines

## Project Overview

JioTV Go is a single Go CLI/server that wraps JioTV Android APIs for live TV, catch-up, EPG, web playback, and IPTV/M3U clients. The server uses Fiber; the web UI is server-rendered HTML with vanilla JavaScript and Tailwind CSS/DaisyUI.

## Architecture & Data Flow

- `main.go` is the only executable. Its `urfave/cli/v2` startup hook loads config, logger, persistent store, and URL encryption before command dispatch.
- `cmd/` owns CLI commands and the Fiber composition root. `cmd/jiotv_go.go:JioTVServer` configures middleware, embedded templates/static files, routes, scheduler, and server listen/TLS.
- `internal/handlers/` owns HTTP behavior. `pkg/` owns JioTV API access, credentials, persistence, EPG generation, URL encryption, and scheduling.
- HLS flow: `/live/...` → token refresh → `pkg/television.Television.Live` → encrypted local render URL → manifest rewrite/proxy handlers (`/render.m3u8`, `/render.ts`, `/render.key`).
- DRM flow: `/play/...` → MPD/license proxy URLs → MPD rewrite → DASH/license proxy handlers. Do not expose or bypass the encrypted upstream-URL flow.
- Runtime state is intentionally global: `config.Cfg`, `handlers.TV`, `utils.Log`, and `store`. Preserve startup order and update shared state through existing lifecycle helpers; do not construct per-request television clients.

## Key Directories

- `cmd/` — CLI implementations, Fiber setup, login, EPG, background process support.
- `internal/config/` — `JioTVConfig` schema and file/environment loading.
- `internal/handlers/` — Fiber handlers for live, catch-up, DRM, EPG, auth, and UI.
- `internal/middleware/` — custom Fiber middleware.
- `internal/utils/` — shared handler error, proxy, URL, and cache helpers.
- `pkg/television/` — concrete `fasthttp` JioTV API client and channel/custom-channel models.
- `pkg/store/` — mutex-protected TOML runtime store; credentials/device state go here.
- `pkg/secureurl/`, `pkg/epg/`, `pkg/scheduler/`, `pkg/utils/` — URL protection, EPG jobs, scheduling, and shared auth/HTTP utilities.
- `web/views/` — embedded Go HTML templates; `web/static/internal/` — browser JS and Tailwind source/output; `web/test/` — Jest tests.
- `configs/` — example YAML/TOML/JSON config files; pass one explicitly with `--config`.
- `scripts/` — release/install utilities, not the regular developer task runner.

## Repository Skills

- `.agents/skills/change-configuration/SKILL.md` — use for any `JIOTV_*` or JSON/YAML/TOML configuration change; it defines the required schema, defaults, examples, consumers, and tests to synchronize.
- `.agents/skills/diagnose-playback/SKILL.md` — use for live, catch-up, HLS, DASH/DRM, browser, or IPTV playback failures; it defines the hop-by-hop evidence and token-scope checks.

## Development Commands

Run from repository root unless stated otherwise:

```bash
# Dependencies and backend build
go mod tidy
go build -o build/jiotv_go .

# Backend tests
go test -v ./...

# Frontend dependencies, generated CSS, and tests
cd web && npm ci
cd web && npm run build
cd web && npm test -- --watchAll=false --ci

# Local server; debug enables template reload and stdout logs
JIOTV_DEBUG=true JIOTV_LOG_TO_STDOUT=true go run main.go serve --host 127.0.0.1 --port 5001

# Built server
./build/jiotv_go serve --host 127.0.0.1 --port 5001
```

Use `go test ./path/to/package -run '^TestName$'` for focused Go work. Use `cd web && npm run test:coverage` for optional frontend coverage. No Makefile or project-local lint command exists.

## Code Conventions & Common Patterns

- Format Go with `gofmt`; use package-local `*_test.go` tests named `TestXxx`. Prefer table cases and `t.Run` where existing tests do.
- Fiber handlers use `func(*fiber.Ctx) error`. Use `internal/utils` JSON error helpers for client errors; propagate/wrap operational errors with `%w` rather than inventing response formats.
- Reuse `pkg/utils` request helpers and `fasthttp` `Acquire*`/`Release*` pairs. Never retain pooled request/response objects.
- Configuration is a tagged global schema: add runtime settings to `internal/config.JioTVConfig` with YAML, JSON, TOML, and `JIOTV_*` tags.
- No dependency-injection/service-interface layer: concrete package types and initialized globals are established patterns. Keep additions consistent unless a real isolation need exists.
- Shared mutable state requires existing synchronization: mutexes for token/store updates, `singleflight` for duplicate refreshes, `sync.Map` plus TTL for caches, and `RWMutex` for custom channels. Preserve these boundaries.
- Frontend is vanilla JS. Follow existing `web/static/internal/*.js` module patterns and Tailwind/DaisyUI utility styling; do not introduce a frontend framework.
- Runtime credentials/data belong under `JIOTV_PATH_PREFIX` (default `$HOME/.jiotv_go`) and must remain untracked. Never commit `.env`, credentials, generated runtime state, or logs.

## Important Files

- `main.go` — CLI entry point and global initialization order.
- `cmd/jiotv_go.go` — Fiber server composition and route registration.
- `internal/config/config.go` — configuration schema/loading.
- `internal/handlers/handlers.go` — handler initialization and HLS/channel behavior.
- `internal/handlers/auth.go` and `internal/handlers/drm.go` — credential refresh and DRM proxy lifecycle.
- `pkg/television/television.go` — JioTV upstream client.
- `pkg/store/store.go` — persistent runtime state.
- `pkg/secureurl/secureurl.go` — upstream URL protection.
- `web/package.json` — npm scripts and frontend dependencies.
- `web/static/internal/input.css` — Tailwind v4/DaisyUI 5 input; `tailwind.css` is its committed generated output.
- `web/jest.config.js` — Jest/jsdom and frontend coverage scope.

## Runtime/Tooling Preferences

- Go module: `github.com/jiotv-go/jiotv_go/v3`.
- Frontend package manager: npm with lockfile v3. Use `npm ci` for reproducible installs. CI uses Node LTS; no repository Node version pin exists.
- Frontend stack is Tailwind CSS 4 with `@tailwindcss/cli` and DaisyUI 5. Configuration is CSS-first in `web/static/internal/input.css`; there is no Tailwind/PostCSS config file.
- `web/static/internal/tailwind.css` is a versioned, minified build artifact. Rebuild and commit it after changing `input.css`, template class usage, or frontend dependencies.
- Docker is optional local tooling: `docker-compose.yml` uses `dev.Dockerfile`, mounts the checkout, reads `.env`, and runs with `JIOTV_DEBUG=true`. Normal local Go development is simpler.
- Do not casually run `scripts/build-binaries.sh`, `scripts/increment-version.sh`, or installers; they are release/end-user tooling and may cross-build, alter `VERSION`, or mutate shell setup.

## Testing & QA

- Backend: standard Go `testing` tests are colocated under `cmd/`, `internal/`, and `pkg/`; Testify `assert` is optional, not required. Keep tests offline and deterministic: use `httptest` or injected clients for upstream APIs, restore mutated globals and caches, and do not run shared-state tests in parallel.
- Frontend: Jest with jsdom; tests live in `web/test/*.test.js` and use DOM/mocked-browser APIs.
- CI quality gate: `go mod tidy`, `go test -v ./...`, then `cd web && npm ci && npm test -- --watchAll=false --ci`.
- Add or update behavior-focused tests for changed contracts. No numeric coverage threshold exists; Jest coverage covers `web/static/internal/**/*.js`, while Go coverage is not configured.
- No ESLint, Prettier, golangci-lint, Staticcheck, or local `go vet` workflow is declared. `.deepsource.toml` configures external Go, Docker, and shell analysis.

## Contribution & Release Workflow

- The default and PR target branch is `develop`; `main` is the release branch. Open normal feature and fix PRs against `develop`.
- Follow Conventional Commits using the repository's established types: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, and `refactor:`. Keep each commit to one logical change.
- Before a PR, run the checks relevant to every changed area. Backend changes require `go test -v ./...`; frontend changes additionally follow `web/AGENTS.md`; configuration and user-facing behavior changes require matching docs/examples.
- Do not manually bump `VERSION`, create release tags, or run release scripts for a normal change. Pushes to `main` trigger `.github/workflows/release.yml`, which updates `develop`, builds release binaries, and publishes tags/releases.

---
name: e2e-contract-testing
description: Syncs a frontend route/selector contract for backend Playwright E2E tests and drives tests using contract keys instead of raw selectors. Use when creating or updating ticketing-system E2E tests, selectors, or routes.
---

# E2E Contract Testing

Use this skill for `ticketing-system/backend/e2e` work where tests should read like behavior steps and rely on stable contract keys.

## Why this setup
- The UI is served by backend (`WithFrontend`), so browser tests run against backend-served frontend.
- Postgres testcontainers provide repeatable isolation for E2E state.
- Contract keys avoid brittle selector duplication in tests.

## What this skill manages
- Frontend E2E contract generation:
  - `ticketing-system/backend/e2e/contracts/frontend_contract.json`
- Contract source map:
  - `ticketing-system/backend/e2e/contracts/frontend_contract.source.json`
- Generator script:
  - `ticketing-system/scripts/generate-e2e-contract.mjs`
- Backend harness and scenario DSL:
  - `ticketing-system/backend/e2e/harness.go`
  - `ticketing-system/backend/e2e/scenario.go`
  - `ticketing-system/backend/e2e/full_flow_test.go`
  - `ticketing-system/backend/e2e/smoke_playwright_test.go`

## Prerequisites
- Docker daemon running.
- Frontend dependencies available for building `frontend/dist`.
- E2E Go dependencies present in `ticketing-system/backend/go.mod`.

## One-time setup
1. Generate frontend E2E contract:
   - `make -C ticketing-system e2e-contract`
2. Add backend E2E dependencies:
   - `cd ticketing-system/backend`
   - `go get github.com/playwright-community/playwright-go@latest github.com/testcontainers/testcontainers-go@latest github.com/testcontainers/testcontainers-go/modules/postgres@latest`
   - `go mod tidy`
3. Install Playwright browser runtime:
   - `cd ticketing-system/backend`
   - `go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium`
4. Build frontend assets:
   - `cd ticketing-system/frontend`
   - `npm ci`
   - `npm run build`

## Default workflow
1. Regenerate contract:
   - Run `bash .codex/skills/e2e-contract-testing/scripts/sync_e2e_contract.sh`
2. If generation fails:
   - Add missing `data-testid` attributes in frontend components.
   - Or update `frontend_contract.source.json` keys to match intended controls.
3. Use contract keys in tests (not raw selectors):
   - Routes via `WhenIGoToRoute(...)`
   - Selectors via `ThenISeeSelectorKey(...)`, `WhenIClickKey(...)`, etc.
4. Keep tests behavior-focused:
   - Prefer `Given/When/Then/And` chains in test files.

## Harness architecture
- Main layers:
  - `Harness` (`backend/e2e/harness.go`): setup/teardown + low-level browser/test helpers.
  - `Scenario` (`backend/e2e/scenario.go`): cucumber-style step chaining wrappers.
  - Tests (`backend/e2e/*_test.go`): behavior-level scenarios using step methods.

### Harness lifecycle (`NewHarness`)
1. Start Postgres testcontainer.
2. Connect store and apply migrations.
3. Seed baseline data (user/project/group/workflow/story).
4. Start backend server with frontend static serving.
5. Load generated frontend contract JSON.
6. Start Playwright runtime and Chromium page.

### Cleanup lifecycle (`Close`)
1. Close page/browser/runtime.
2. Stop server/store/container.
3. Finalize context cancellation.

## Harness API reference

### Contract and routing
- `GoToRoute(routeKey, params)`:
  - Navigate using route keys from contract, not hardcoded paths.
- `ResolveRoute(routeKey, params)`:
  - Resolve and validate route params.
- `Selector(selectorKey)`:
  - Resolve selector key to concrete CSS selector.

### Interaction helpers
- `WaitVisible`, `WaitVisibleKey`
- `WaitHidden`
- `Click`, `ClickKey`
- `Fill`, `FillKey`
- `Press`, `PressKey`
- `SelectOptionByValue`, `SelectOptionByValueKey`
- `ExpectTextVisible`
- `ExpectURLContains`

### Test data helper
- `SeedData()` returns seeded IDs (project/story) for deterministic flows.

## Scenario DSL reference
- Core combinators:
  - `Given(...)`, `When(...)`, `Then(...)`, `And(...)`
- Common helpers:
  - `GivenAppIsRunning()`
  - `WhenIGoTo(...)`, `WhenIGoToRoute(...)`
  - `WhenIClick(...)`, `WhenIClickKey(...)`
  - `WhenIFill(...)`, `WhenIFillKey(...)`
  - `WhenIPress(...)`, `WhenIPressKey(...)`
  - `WhenILogInAs(...)`
  - `WhenISelectProjectByID(...)`
  - `ThenISeeSelector(...)`, `ThenISeeSelectorKey(...)`
  - `ThenISeeText(...)`
  - `ThenURLContains(...)`
  - `AndISeeSelector(...)`, `AndISeeSelectorKey(...)`
  - `AndISeeText(...)`

## Failure handling
- Each step logs `Given/When/Then/And` text.
- On failure the harness captures:
  - current URL
  - full-page screenshot
  - HTML snapshot
- Artifacts default to:
  - `ticketing-system/backend/e2e/artifacts/<test-name>/...`

## Rules
- Do not hardcode new selectors in test files when a contract key should exist.
- Add/edit keys in `frontend_contract.source.json`, then regenerate contract.
- Keep key names semantic and stable (`login.submit_button`, `nav.logout_button`).
- Prefer `data-testid` selectors for long-term stability over CSS class selectors.
- Prefer scenario step methods over direct Playwright calls in tests.
- Use seeded IDs from `scenario.SeedData()` for deterministic full-flow tests.

## Writing tests
- Start each test with:
  - `scenario := NewScenario(t)`
  - `defer scenario.Close()`
- Use route/selector keys from `frontend_contract.json`:
  - `WhenIGoToRoute("home")`
  - `ThenISeeSelectorKey("login.view")`
- Use high-level flow steps where available:
  - `WhenILogInAs("AdminUser", "admin123")`
  - `WhenISelectProjectByID(seed.ProjectID)`

Example:
```go
func TestLoginScreen(t *testing.T) {
  scenario := NewScenario(t)
  defer scenario.Close()

  scenario.
    GivenAppIsRunning().
    WhenIGoToRoute("home").
    ThenISeeSelectorKey("login.view")
}
```

Full-flow pattern:
```go
seed := scenario.SeedData()
scenario.
  GivenAppIsRunning().
  WhenIGoToRoute("home").
  WhenILogInAs("AdminUser", "admin123").
  WhenISelectProjectByID(seed.ProjectID).
  ThenURLContains("/projects/" + seed.ProjectID + "/board")
```

## Runtime configuration
- `E2E_FRONTEND_DIR`: custom frontend dist path.
- `E2E_FRONTEND_CONTRACT`: custom contract JSON path.
- `E2E_TEST_TIMEOUT`: total test timeout (default `3m`).
- `E2E_STEP_TIMEOUT`: action/selector timeout (default `15s`).
- `E2E_NAV_TIMEOUT`: navigation timeout (default `20s`).
- `E2E_HEADLESS`: browser mode (`true`/`false`, default `true`).
- `E2E_ARTIFACTS_DIR`: failure artifact output directory.
- `E2E_MAX_PARALLEL`: max concurrent harness instances (default `2`).
  - Use lower values if Docker/Playwright contention causes flaky timeouts.

## Quick commands
- Regenerate contract:
  - `make -C ticketing-system e2e-contract`
- Run E2E:
  - `make -C ticketing-system e2e`
- Compile-check E2E package:
  - `cd ticketing-system/backend && GOCACHE=/tmp/go-build go test -tags=e2e ./e2e -c`

# CLAUDE.md — paymail-inspector agent guide

> **Read first:** [AGENTS.md](AGENTS.md) is the root authority for coding,
> commit, PR, release, and CI rules. [tech-conventions/](tech-conventions/)
> has the deep guides. If anything below conflicts with those, **they win.**

## What this project is

`paymail-inspector` is a CLI for inspecting paymail service providers: it
discovers a domain's capabilities, resolves and verifies paymail addresses,
validates a domain against the bsvalias specification (SRV, DNSSEC, SSL,
capabilities), searches public providers for a handle, and lists, searches,
and generates BRFC IDs. The binary also answers to `paymail` when invoked
through a symlink of that name.

- **Module:** `github.com/mrz1836/paymail-inspector`. Go 1.26, CGO=0, no `/vendor`.
- **Layout:**
  - [../cmd/paymail-inspector/](../cmd/paymail-inspector/): the `main` shim that carries ldflags build metadata.
  - [../internal/cli/](../internal/cli/): the cobra command tree. It is built fresh per run, with no mutable globals; dependencies hang off `*app`.
  - [../internal/cache/](../internal/cache/): a file-backed TTL response cache, written atomically on exit.
  - [../internal/output/](../internal/output/): the styled terminal printer, which owns the color decision.
- **Paymail protocol:** [go-paymail](https://github.com/bsv-blockchain/go-paymail).
  The CLI depends only on the narrow `paymailClient` interface in `internal/cli/app.go`.

## Build, lint, test: always via `magex`

Full reference: [tech-conventions/mage-x.md](tech-conventions/mage-x.md).

| Command            | When to use                                       |
|--------------------|---------------------------------------------------|
| `magex format:fix` | Before every commit                               |
| `magex lint`       | All linters (`magex lint:fix` to auto-fix)        |
| `magex test`       | Fast unit tests (default development loop)        |
| `magex test:race`  | Race detector: required (`whois` is concurrent) |
| `magex test:fuzz`  | Fuzz the input parsers                            |
| `magex build`      | Build `cmd/paymail-inspector/paymail-inspector`   |

Coverage target: **≥ 90%** (enforced in CI via `.github/env/90-project.env`).

## Testing rules

- **Tests never touch the network.** Commands get a fake `paymailClient`
  (`internal/cli/helpers_test.go`). Self-update tests use a fake release source
  and an `httptest` server. `TestMain` sets `NO_UPDATE_CHECK=1`.
- Use the `harness` helper: temp config and cache paths, captured stdout and
  stderr, and exit codes (`ExitOK`, `ExitError`, `ExitUsage`).
- Third-party integrations that shut down (Bitpic, Baemail, Roundesk, PowPing,
  Dime.ly) were removed. Their old `--skip-*` flags remain as hidden,
  deprecated no-ops. Don't re-add integrations without a live, verified API.

## Don't

- Don't run raw `go test` / `go build` / `golangci-lint` for CI decisions. Use `magex`.
- Don't tag releases. CI publishes them from `magex version:bump bump=… push`.
- Don't edit files synced by go-broadcast (workflows, actions, env `00`–`20`,
  tech-conventions, `.goreleaser.yml`, `.mage.yaml`, …). Change them upstream.
  The project-specific files are `90-project.env`, `homebrew-tap.yml`,
  `dependabot.yml`, `.golangci.json`, `.gitignore` and this file.
- Don't use `--no-verify` on commits.
- Don't author `tech-conventions/` content into this file. Link to it.

<div align="center">

# 🔎&nbsp;&nbsp;Paymail Inspector

**Inspect, validate, and resolve paymail addresses from your terminal**

Capability discovery · address resolution · spec validation · handle lookup · BRFC tooling

<br/>

<a href="https://github.com/mrz1836/paymail-inspector/releases"><img src="https://img.shields.io/github/release-pre/mrz1836/paymail-inspector?include_prereleases&style=flat-square&logo=github&color=black" alt="Release"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/github/go-mod/go-version/mrz1836/paymail-inspector?style=flat-square&logo=go&color=00ADD8" alt="Go Version"></a>
<a href="https://github.com/mrz1836/paymail-inspector/blob/master/LICENSE"><img src="https://img.shields.io/github/license/mrz1836/paymail-inspector?style=flat-square&color=blue&v=2" alt="License"></a>

<br/>

<table align="center" border="0">
  <tr>
    <td align="right">
       <code>CI / CD</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/paymail-inspector/actions"><img src="https://img.shields.io/github/actions/workflow/status/mrz1836/paymail-inspector/fortress.yml?branch=master&label=build&logo=github&style=flat-square" alt="Build"></a>
       <a href="https://github.com/mrz1836/paymail-inspector/actions"><img src="https://img.shields.io/github/last-commit/mrz1836/paymail-inspector?style=flat-square&logo=git&logoColor=white&label=last%20update" alt="Last Commit"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Quality</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://codecov.io/gh/mrz1836/paymail-inspector"><img src="https://codecov.io/gh/mrz1836/paymail-inspector/branch/master/graph/badge.svg?style=flat-square" alt="Coverage"></a>
    </td>
  </tr>

  <tr>
    <td align="right">
       <code>Security</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://scorecard.dev/viewer/?uri=github.com/mrz1836/paymail-inspector"><img src="https://api.scorecard.dev/projects/github.com/mrz1836/paymail-inspector/badge?style=flat-square" alt="Scorecard"></a>
       <a href=".github/SECURITY.md"><img src="https://img.shields.io/badge/policy-active-success?style=flat-square&logo=security&logoColor=white" alt="Security"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Community</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/paymail-inspector/graphs/contributors"><img src="https://img.shields.io/github/contributors/mrz1836/paymail-inspector?style=flat-square&color=orange" alt="Contributors"></a>
       <a href="https://mrz1818.com/"><img src="https://img.shields.io/badge/donate-bitcoin-ff9900?style=flat-square&logo=bitcoin" alt="Bitcoin"></a>
    </td>
  </tr>
</table>

</div>

<br/>
<br/>

<div align="center">

### <code>Project Navigation</code>

</div>

<table align="center">
  <tr>
    <td align="center" width="33%">
       🚀&nbsp;<a href="#-installation"><code>Installation</code></a>
    </td>
    <td align="center" width="33%">
       ⚡&nbsp;<a href="#-quick-start"><code>Quick&nbsp;Start</code></a>
    </td>
    <td align="center" width="33%">
       📚&nbsp;<a href="#-documentation"><code>Documentation</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       🔐&nbsp;<a href="#-security"><code>Security</code></a>
    </td>
    <td align="center">
      🛠️&nbsp;<a href="#-code-standards"><code>Code&nbsp;Standards</code></a>
    </td>
    <td align="center">
      🧪&nbsp;<a href="#-examples--tests"><code>Examples&nbsp;&&nbsp;Tests</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
      🤖&nbsp;<a href="#-ai-usage--assistant-guidelines"><code>AI&nbsp;Usage</code></a>
    </td>
    <td align="center">
       ⚖️&nbsp;<a href="#-license"><code>License</code></a>
    </td>
    <td align="center">
       🤝&nbsp;<a href="#-contributing"><code>Contributing</code></a>
    </td>
  </tr>
  <tr>
    <td align="center" colspan="3">
       👥&nbsp;<a href="#-maintainers"><code>Maintainers</code></a>
    </td>
  </tr>
</table>

<br/>

### 🎥 Demo

<img src="examples/demo.gif" alt="Paymail Inspector Demo" title="Paymail Inspector Demo"/>

<br/>

**Paymail Inspector is a single Go binary for anyone who builds, runs, or debugs
[paymail](https://tsc.bsvblockchain.org/standards/paymail/) services.** Point it at a domain
or an address and it shows the whole protocol at work: the SRV record that picks the host,
the capabilities the host publishes, the public key and payment destination an address
resolves to, and the network timing of every request.

It answers to two names: `paymail-inspector` and `paymail`, a shortcut that every install
method below sets up. The examples use the short one.

<br/>

## 🚀 Installation

**Paymail Inspector** ships as a single static binary (`CGO_ENABLED=0`) for Linux, macOS, and Windows
on amd64 and arm64.

Install the latest prebuilt release into `~/.local/bin`, a user-writable directory, so there's
no `sudo` and `paymail update` can self-update in place afterward.

No `curl … | sudo bash` here: we don't ask you to pipe a mystery script into your shell
and hope for the best. Every line below is in the open, and the download is checked against
the release's published SHA‑256 checksums before anything lands on your `PATH`:

```bash
# Install the latest paymail-inspector release into ~/.local/bin, verified against checksums.txt
VER=$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/mrz1836/paymail-inspector/releases/latest | sed 's#.*/v##')
OS=$(uname -s | tr '[:upper:]' '[:lower:]'); ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
F="paymail-inspector_${VER}_${OS}_${ARCH}.tar.gz"; U="https://github.com/mrz1836/paymail-inspector/releases/download/v${VER}"
mkdir -p ~/.local/bin && cd "$(mktemp -d)" && curl -fsSLO "$U/$F" \
  && WANT=$(curl -fsSL "$U/paymail-inspector_${VER}_checksums.txt" | awk -v f="$F" '$2==f{print $1}') \
  && GOT=$( { command -v sha256sum >/dev/null && sha256sum "$F" || shasum -a 256 "$F"; } | awk '{print $1}') \
  && [ -n "$WANT" ] && [ "$WANT" = "$GOT" ] \
  && tar -xzf "$F" -C ~/.local/bin paymail-inspector \
  && ln -sf paymail-inspector ~/.local/bin/paymail
paymail --version
```

If `paymail` isn't found afterward, add `~/.local/bin` to your `PATH` (put
`export PATH="$HOME/.local/bin:$PATH"` in your `~/.zshrc` or `~/.bashrc`).

<details>
<summary><strong>Homebrew (macOS & Linux)</strong></summary>
<br/>

```bash
brew install mrz1836/paymail-inspector/paymail-inspector
paymail --version
```

The formula installs `paymail-inspector` along with the `paymail` shortcut. Upgrade it with
`brew upgrade paymail-inspector`. `paymail update` leaves Homebrew-owned binaries alone.

</details>

<details>
<summary><strong>Go install / build from source (contributors)</strong></summary>
<br/>

Requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy) (Go 1.26+).

```bash
go install github.com/mrz1836/paymail-inspector/cmd/paymail-inspector@latest
```

Or build from a checkout with [MAGE-X](https://github.com/mrz1836/mage-x):

```bash
git clone https://github.com/mrz1836/paymail-inspector.git && cd paymail-inspector
magex build && install -m 0755 cmd/paymail-inspector/paymail-inspector ~/.local/bin/paymail-inspector
ln -sf paymail-inspector ~/.local/bin/paymail
```

</details>

<br/>

### Keep Paymail Inspector up to date

`paymail update` (alias `paymail upgrade`) downloads the latest release, verifies its
SHA‑256 checksum against the published `paymail-inspector_<ver>_checksums.txt`, and atomically
replaces the binary. It follows the `paymail` symlink to the real file, so the shortcut keeps
working.

```bash
paymail update            # download & install the latest release
paymail update --check    # report whether a newer version is available
paymail update --force    # reinstall the latest even if already current
paymail update --verbose  # narrate each step
```

Every other command also runs a passive, cached background check and prints a one‑line
"a new version is available" notice. It never blocks or fails a command, is skipped for
development builds, and is silenced by `PAYMAIL_INSPECTOR_NO_UPDATE_CHECK=1` (or the shared
`NO_UPDATE_CHECK` / `CI`). If you hit GitHub API rate limits, a token is read from
`PAYMAIL_INSPECTOR_GITHUB_TOKEN`, then `GITHUB_TOKEN`, then `GH_TOKEN`.

> **Heads up:** `paymail update` **refuses** to overwrite a binary that another installer owns,
> such as one in `go install`'s `~/go/bin` or under a Homebrew prefix, because that would break the
> tool that owns it. Upgrade those with their own installer.

<br/>

## 🗺️ Command palette

Every command at a glance. Each one works against live providers.

| Command                                           | Purpose                                                                         |
|---------------------------------------------------|---------------------------------------------------------------------------------|
| `paymail capabilities <domain\|paymail>`          | Discover the host (SRV record) and list the capabilities it publishes           |
| `paymail resolve <paymail>`                       | Resolve an address to its pubkey, output script, address, and public profile    |
| `paymail p2p <paymail>`                           | Start a P2P payment destination request and list its outputs                    |
| `paymail validate <domain\|paymail>`              | Check SRV, DNSSEC, SSL, and required capabilities; exits non-zero on failure    |
| `paymail verify <paymail> <pubkey>`               | Ask the provider whether a pubkey belongs to an address                         |
| `paymail whois <handle>`                          | Search every public provider for a handle, in parallel                          |
| `paymail brfc list` / `search` / `generate`       | Explore BRFC specifications and generate new BRFC IDs                           |
| `paymail update` / `version`                      | Self-update from a GitHub release; show build information                       |

Global flags work on every command: `--config <path>`, `--no-cache`, `--flush-cache`,
`-t/--skip-tracing`, `--no-color`, and `--bsvalias <version>`.

Handles work wherever an address does: `$handle` is a HandCash address and `1handle` is a RelayX
address (quote or escape the `$` in your shell: `\$handcash`).

<br/>

## ⚡ Quick Start

<br/>

### Discover a provider's capabilities

```bash
paymail capabilities handcash.io
```

Follows the `_bsvalias._tcp` SRV record to the real host, then lists every capability with
the BRFC specification it implements.

<br/>

### Resolve an address

```bash
paymail resolve mrz@relayx.io
```

Shows the pubkey, the output script, the address, and the public profile. Request tracing
(DNS, TLS, and time-to-first-byte per call) is on by default; `-t` hides it.

<br/>

### Validate your paymail setup

```bash
paymail validate yourdomain.com
paymail validate you@yourdomain.com --nameserver 1.1.1.1
```

Runs the bsvalias checks (SRV, DNSSEC, SSL, required capabilities, and the pubkey for an
address). A failed required check exits non-zero, so you can gate a deploy on it.

<br/>

### Find who owns a handle

```bash
paymail whois mrz
```

Searches HandCash, RelayX, Simply Cash, Paymail.us, and 1Sat Ordinals in parallel, and says
where the handle is taken and where it is still free.

<br/>

### Verify a pubkey or request P2P outputs

```bash
paymail verify mrz@simply.cash 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec
paymail p2p \$handcash --satoshis 5000
```

<br/>

### Work with BRFC specifications

```bash
paymail brfc list
paymail brfc search nChain
paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```

<br/>

> 📖 **For real output from every command, see the [Examples →](docs/examples.md)**

<br/>

## 📚 Documentation

| Document                                     | Description                                                       |
|----------------------------------------------|-------------------------------------------------------------------|
| **[Examples](docs/examples.md)**             | Every command with real output from live providers                |
| **[Command reference](docs/commands/paymail.md)** | Generated flag and usage reference for each command          |
| **[Example config](config-example.yaml)**    | Every configuration key, with its flag and environment variable   |

<br/>

### Implemented BRFC specifications

| Specification                                                                                                          | BRFC ID        | Used by                          |
|------------------------------------------------------------------------------------------------------------------------|----------------|----------------------------------|
| [BRFC ID Assignment](https://bitcoin-sv-specs.github.io/brfc-paymail/01-02-brfc-id-assignment.html)                    | (algorithm)    | `brfc generate`, `brfc list`     |
| [Service Discovery](https://bitcoin-sv-specs.github.io/brfc-paymail/02-service-discovery.html) (SRV + capabilities)    | `b2aa66e26b43` | every command                    |
| [Public Key Infrastructure](https://bitcoin-sv-specs.github.io/brfc-paymail/03-public-key-infrastructure.html)         | `0c4339ef99c2` | `resolve`, `validate`, `whois`   |
| [Basic Address Resolution](https://bitcoin-sv-specs.github.io/brfc-paymail/04-01-basic-address-resolution.html)        | `759684b1a19a` | `resolve`                        |
| [Sender Validation](https://bitcoin-sv-specs.github.io/brfc-paymail/04-02-sender-validation.html)                      | `6745385c3fc0` | `resolve`                        |
| [Verify Public Key Owner](https://bitcoin-sv-specs.github.io/brfc-paymail/05-verify-public-key-owner.html)             | `a9f510c16bde` | `verify`                         |
| [Public Profile](https://github.com/bitcoin-sv-specs/brfc-paymail/pull/7/files) (name & avatar)                        | `f12f968c92d6` | `resolve`, `p2p`, `whois`        |
| [P2P Payment Destination](https://tsc.bsvblockchain.org/standards/paymail/)                                            | `2a40af698840` | `p2p`                            |

`capabilities` lists every other capability a provider publishes (P2P transactions, BEEF,
PIKE, and so on), named after its BRFC specification when the ID is known.

<br/>

<details>
<summary><strong><code>Public Paymail Providers</code></strong></summary>
<br/>

`whois` searches these providers. Each one was checked live (SRV record and
`/.well-known/bsvalias`). Providers that have shut down are removed rather than left to time out.

| Provider                                 | Domain        |
|------------------------------------------|---------------|
| [HandCash](https://handcash.io)          | `handcash.io` |
| [RelayX](https://relayx.com)             | `relayx.io`   |
| [Simply Cash](https://simply.cash)       | `simply.cash` |
| [Paymail.us](https://paymail.us)         | `paymail.us`  |
| [1Sat Ordinals](https://1sat.app)        | `1sat.app`    |

</details>

<details>
<summary><strong><code>Configuration</code></strong></summary>
<br/>

Settings are read in this order: flags, then `PAYMAIL_INSPECTOR_*` environment variables, then the
config file, then the defaults. The config file is optional. It lives in your user config directory:

| OS      | Path                                                         |
|---------|--------------------------------------------------------------|
| macOS   | `~/Library/Application Support/paymail-inspector/config.yaml` |
| Linux   | `~/.config/paymail-inspector/config.yaml`                    |
| Windows | `%AppData%\paymail-inspector\config.yaml`                    |

A config file from an older release at `~/paymail/config.yaml` is still read. Use `--config` to
point at any other file. See the [example config](config-example.yaml) for every key.

```bash
export PAYMAIL_INSPECTOR_SENDER_HANDLE=you@yourdomain.com   # same as --sender-handle
```

</details>

<details>
<summary><strong><code>Response Cache</code></strong></summary>
<br/>

Capabilities, SRV records, public keys, and public profiles are cached for an hour in your user
cache directory (`<user cache dir>/paymail-inspector/cache.json`). This keeps repeated commands fast
and spares the providers. Address resolution, P2P requests, and verification always go to the provider.

```bash
paymail resolve mrz@relayx.io --no-cache   # skip cached responses for one command
paymail --flush-cache                      # empty the cache
```

The cache is a single JSON file that is replaced atomically, with no background processes or locks,
so several terminals can run Paymail Inspector at once. An unreadable cache is discarded and never
fails a command.

</details>

<details>
<summary><strong><code>Retired Integrations</code></strong></summary>
<br/>

Earlier releases also queried Bitpic, Baemail, Roundesk, PowPing, and Dime.ly. Those services have
shut down, so the integrations are gone. Their `--skip-*` flags are still accepted as hidden no-ops
that print a deprecation notice, so existing scripts keep working.

</details>

<details>
<summary><strong><code>Package Dependencies</code></strong></summary>
<br/>

- [go-paymail](https://github.com/bsv-blockchain/go-paymail) for the paymail protocol, DNS, and SSL checks
- [cobra](https://github.com/spf13/cobra) and [viper](https://github.com/spf13/viper) for the CLI and configuration
- [color](https://github.com/fatih/color) for terminal colors
- [go-sanitize](https://github.com/mrz1836/go-sanitize) for input sanitation
- [go-selfupdate](https://github.com/mrz1836/go-selfupdate) for checksum-verified self-updates
- [resty](https://github.com/go-resty/resty) for HTTP request tracing (through go-paymail)

</details>

<br/>

## 🔐 Security

Paymail Inspector only **reads** public paymail data. It never holds private keys or signs
anything, and it sends nothing but standard paymail requests to the providers you name. Two
requests are worth knowing about:

- `p2p` asks the receiver to **create** a payment destination. That is what the protocol does, but
  the provider records the request.
- When a provider enforces sender validation and you don't pass `--signature`, `resolve` sends a
  random placeholder signature so the rest of the flow can be inspected. A strict provider rejects it.

For security issues, see our [Security Policy](.github/SECURITY.md) or contact:
[paymail-inspector@mrz1818.com](mailto:paymail-inspector@mrz1818.com).

<br/>

### Additional Documentation & Repository Management

<details>
<summary><strong><code>Development Setup (Getting Started)</code></strong></summary>
<br/>

Install the [MAGE-X](https://github.com/mrz1836/mage-x) build tool for development:

```bash
# Install MAGE-X for development and building
go install github.com/mrz1836/mage-x/cmd/magex@latest
magex update:install
```

Set up the [go-pre-commit](https://github.com/mrz1836/go-pre-commit) hooks, which run the same
formatting, linting, and secret scanning as CI before every commit:

```bash
go install github.com/mrz1836/go-pre-commit/cmd/go-pre-commit@latest
go-pre-commit install
```

</details>

<details>
<summary><strong><code>Build Commands</code></strong></summary>
<br/>

View all build commands:

```bash
magex help
```

Common commands:
- `magex build`: build the binary
- `magex test`: run the test suite
- `magex lint`: run all linters
- `magex deps:update`: update dependencies

Regenerate the [command reference](docs/commands/paymail.md) after changing a command:

```bash
go run ./cmd/paymail-inspector gen-docs
```

Re-record the [demo](examples/demo.gif) with [VHS](https://github.com/charmbracelet/vhs) (needs network access):

```bash
vhs examples/demo.tape
```

</details>

<details>
<summary><strong><code>Binary Deployment</code></strong></summary>
<br/>

This project uses [goreleaser](https://github.com/goreleaser/goreleaser) for
streamlined binary deployment to GitHub. To get started, install it via:

```bash
brew install goreleaser
```

The release process is defined in the [.goreleaser.yml](.goreleaser.yml)
configuration file. Then create and push a new Git tag using:

```bash
magex version:bump bump=patch push=true branch=master
```

The tag triggers the Fortress release. Once it succeeds, the
[Homebrew tap workflow](.github/workflows/homebrew-tap.yml) publishes the new version to
[mrz1836/homebrew-paymail-inspector](https://github.com/mrz1836/homebrew-paymail-inspector)
using a short-lived token from the repository's auto-merge GitHub App, so there is no
personal access token to expire.

</details>

<details>
<summary><strong><code>GitHub Workflows</code></strong></summary>
<br/>

Paymail Inspector uses the **Fortress** workflow system for comprehensive CI/CD:

- **fortress-test-suite.yml**: complete test suite across multiple Go versions
- **fortress-code-quality.yml**: code quality checks (gofmt, golangci-lint, staticcheck)
- **fortress-security-scans.yml**: security vulnerability scanning
- **fortress-coverage.yml**: code coverage reporting to Codecov
- **fortress-release.yml**: automated binary releases via GoReleaser
- **homebrew-tap.yml**: publishes each release to the Homebrew tap

All workflows are driven by modular configuration in [`.github/env/`](.github/env/README.md).
**[View all workflows and the control center →](.github/docs/workflows.md)**

</details>

<details>
<summary><strong><code>Repository Features</code></strong></summary>
<br/>

This repository includes 25+ built-in features covering CI/CD, security, code quality, developer
experience, and community tooling.

**[View the full Repository Features list →](.github/docs/repository-features.md)**

</details>

<details>
<summary><strong><code>Updating Dependencies</code></strong></summary>
<br/>

To update all dependencies (Go modules, linters, and related tools), run:

```bash
magex deps:update
```

This command ensures all dependencies are brought up to date in a single step,
including Go modules and any managed tools. It is the recommended way to keep
your development environment and CI in sync with the latest versions.

</details>

<br/>

## 🧪 Examples & Tests

All unit tests run via [GitHub Actions](https://github.com/mrz1836/paymail-inspector/actions) and use
[Go version 1.26](https://go.dev/doc/go1.26). View the [configuration file](.github/workflows/fortress.yml).

The suite runs **fully offline**. Commands talk to a fake paymail client, and self-update tests use a
local test server, so every run is fast and repeatable and never depends on a provider being up.

Run all tests (fast):

```bash
magex test
```

Run all tests with race detector (slower):

```bash
magex test:race
```

Fuzz the input parsers:

```bash
magex test:fuzz
```

<br/>

### Test Coverage

View coverage report:

```bash
magex test:coverage
```

Coverage reports are automatically uploaded to [Codecov](https://codecov.io/gh/mrz1836/paymail-inspector)
on every commit, and CI fails below 90%.

<br/>

## 🛠️ Code Standards

Read more about this Go project's [code standards](.github/CODE_STANDARDS.md).

<br/>

## 🤖 AI Usage & Assistant Guidelines

Read the [AI Usage & Assistant Guidelines](.github/CLAUDE.md) for details on
how AI is used in this project and how to interact with AI assistants.

<br/>

## 👥 Maintainers

| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) | [<img src="https://github.com/rohenaz.png" height="50" alt="Satchmo" />](https://github.com/rohenaz) |
|:------------------------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------------:|
|                                [MrZ](https://github.com/mrz1836)                                 |                                [Satchmo](https://github.com/rohenaz)                                 |

<br/>

## 🤝 Contributing

View the [contributing guidelines](.github/CONTRIBUTING.md) and please follow
the [code of conduct](.github/CODE_OF_CONDUCT.md).

### How can I help?

All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to
raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/mrz1836) :clap:
or by making a [**bitcoin donation**](https://mrz1818.com/?tab=tips&utm_source=github&utm_medium=sponsor-link&utm_campaign=paymail-inspector&utm_term=paymail-inspector&utm_content=paymail-inspector)
to ensure this journey continues indefinitely! :rocket:

[![Stars](https://img.shields.io/github/stars/mrz1836/paymail-inspector?label=Please%20like%20us&style=social)](https://github.com/mrz1836/paymail-inspector/stargazers)

### Credits

Inspiration and code snippets from [dnssec](https://github.com/binaryfigments/dnssec) and
[check-ssl](https://github.com/wycore/check-ssl). The demo is recorded with
[VHS](https://github.com/charmbracelet/vhs).

<br/>

## 📝 License

[![License](https://img.shields.io/github/license/mrz1836/paymail-inspector.svg?style=flat&v=2)](LICENSE)

This project is licensed under the terms of the [`LICENSE`](LICENSE) file at
the repo root.

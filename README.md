# Paymail Inspector

> CLI application for interacting with paymail service providers

[![Release](https://img.shields.io/github/release-pre/mrz1836/paymail-inspector.svg?logo=github&style=flat&v=7)](https://github.com/mrz1836/paymail-inspector/releases)
[![Downloads](https://img.shields.io/github/downloads/mrz1836/paymail-inspector/total.svg?logo=github&style=flat&v=7)](https://github.com/mrz1836/paymail-inspector/releases)
[![Build Status](https://travis-ci.com/mrz1836/paymail-inspector.svg?branch=master&v=7)](https://travis-ci.com/mrz1836/paymail-inspector)
[![Report](https://goreportcard.com/badge/github.com/mrz1836/paymail-inspector?style=flat&v=7)](https://goreportcard.com/report/github.com/mrz1836/paymail-inspector)
[![Go](https://img.shields.io/github/go-mod/go-version/mrz1836/paymail-inspector?v=7)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-MrZ-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/mrz1836)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat)](https://mrz1818.com/?tab=tips&af=paymail-inspector)

<br/>

<img src=".github/IMAGES/paymail-inspector.gif?raw=true&v=7" alt="Paymail Commands">

<br/>

## Table of Contents
- [Installation](#installation)
- [Commands](#commands)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Code Standards](#code-standards)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

<br/>

## Installation

**Install with [brew](https://github.com/mrz1836/homebrew-paymail-inspector)**
```shell script
brew tap mrz1836/paymail-inspector && brew install paymail-inspector
paymail
```

**Install using a [compiled binary](https://github.com/mrz1836/paymail-inspector/releases)** on Linux, Mac or Windows _(Mac example)_
```shell script
curl -LkSs https://github.com/mrz1836/paymail-inspector/releases/download/v0.1.12/paymail-inspector_macOS_64-bit.tar.gz -o app.tar.gz
tar -zxf app.tar.gz && cd ./app/
./paymail
```

**Install with [go](https://formulae.brew.sh/formula/go)**
```shell script
go get github.com/mrz1836/paymail-inspector
cd /$GOPATH/src/github.com/mrz1836/paymail-inspector && make install
paymail
```

<br/>

## Commands

### `brfc`
> List all known brfc specifications ([view example](docs/examples.md#list-brfc-specifications))
```shell script
paymail brfc list
```

<br/>

> Generate a new `BRFC ID` for a new specification ([view example](docs/examples.md#generate-new-brfc-id))
```shell script
paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```
 
<br/>

> Search all brfc specifications (id, title, author) ([view example](docs/examples.md#search-brfc-specifications))
```shell script
paymail brfc search nChain
```

<br/>

___

<br/>

### `capabilities`
> Lists the available capabilities of the paymail service ([view example](docs/examples.md#get-capabilities-by-domain))
```shell script
paymail capabilities moneybutton.com
```

<br/>

___

<br/>

### `p2p`
> Starts a P2P payment request and returns (n) outputs of (`script`,`satoshis`,`address`) ([view example](docs/examples.md#start-p2p-payment-request-by-paymail))
```shell script
paymail p2p mrz@moneybutton.com
```

<br/>

___

<br/>

### `resolve`
> Returns the `pubkey`, `output script`, `address` and `profile` for a given paymail address ([view example](docs/examples.md#resolve-paymail-address-by-paymail))
```shell script
paymail resolve mrz@moneybutton.com
```

<br/>

___

<br/>

### `validate`
> Runs several validations on the paymail service for DNSSEC, SSL, SRV and required capabilities ([view example](docs/examples.md#validate-paymail-setup-by-paymail-or-domain))
```shell script
paymail validate moneybutton.com
```

<br/>

___

<br/>

### `verify`
> Verifies if a paymail is associated to a pubkey ([view example](docs/examples.md#verify-public-key-owner))
```shell script
paymail verify mrz@moneybutton.com 02ead23149a1e33df17325ec7a7ba9e0b20c674c57c630f527d69b866aa9b65b10
``` 

<br/>

___

<br/>

### `whois`
> Searches all public paymail providers for a given handle ([view example](docs/examples.md#whois-for-handles))
```shell script
paymail whois mrz
```

<br/>

## Documentation
Get started with the [examples](docs/examples.md). View the generated golang [godocs](https://pkg.go.dev/github.com/mrz1836/paymail-inspector?tab=subdirectories).

All the generated command documentation can be found in [docs/commands](docs/commands).

This application was built using the [official paymail specifications](http://bsvalias.org/index.html).

Additional paymail information can also be found via [MoneyButton's documentation](https://docs.moneybutton.com/docs/paymail-overview.html).

### Implemented [BRFCs](http://bsvalias.org/01-brfc-specifications.html)
- [x] BRFC ID Assignment ([assignment](http://bsvalias.org/01-02-brfc-id-assignment.html))
- [x] Service Discovery ([b2aa66e26b43](http://bsvalias.org/02-service-discovery.html))
- [x] Public Key Infrastructure (pki) ([0c4339ef99c2](http://bsvalias.org/03-public-key-infrastructure.html))
- [x] Basic Address Resolution ([759684b1a19a](http://bsvalias.org/04-01-basic-address-resolution.html))
- [x] Verify Public Key Owner ([a9f510c16bde](http://bsvalias.org/05-verify-public-key-owner.html))
- [x] PayTo Protocol Prefix ([7bd25e5a1fc6](http://bsvalias.org/04-04-payto-protocol-prefix.html))
- [x] Public Profile [(f12f968c92d6)](https://github.com/bitcoin-sv-specs/brfc-paymail/pull/7/files)
- [x] P2P Payment Destination ([2a40af698840](https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html))
- [x] Sender Validation ([6745385c3fc0](http://bsvalias.org/04-02-sender-validation.html))
- [ ] P2P Transactions ([5f1323cddf31](https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html))
- [ ] Receiver Approvals ([3d7c2ca83a46](http://bsvalias.org/04-03-receiver-approvals.html))
- [ ] Merchant API ([ce852c4c2cd1](https://github.com/bitcoin-sv-specs/brfc-merchantapi))
- [ ] JSON Envelope Specification ([298e080a4598](https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/jsonenvelope))
- [ ] Fee Specification ([fb567267440a](https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/feespec))
- [ ] MinerID ([07f0786cdab6](https://github.com/bitcoin-sv-specs/brfc-minerid))
- [ ] MinerID Extension: FeeSpec ([62b21572ca46](https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/feespec))
- [ ] MinerID Extension: MinerParams ([1b1d980b5b72](https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/minerparams))
- [ ] MinerID Extension: BlockInfo ([a224052ad433](https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/blockinfo))
- [ ] MinerID Extension: BlockBind ([b8930c2bbf5d](https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/blockbind))

<details>
<summary><strong><code>Public Paymail Providers</code></strong></summary>
<br/>

- [MoneyButton](https://tncpw.co/4c58a26f)
- [Handcash](https://tncpw.co/742b1f09)
- [RelayX](https://tncpw.co/4897634e)
- [Centbee](https://tncpw.co/4350c72f)
- [Simply.cash](https://tncpw.co/1ce8f70f)
- [Bit.sv](https://tncpw.co/bff8984c)
- [myPaymail](https://tncpw.co/ee243a15)
- [Volt](https://tncpw.co/e9ff2b0c)
</details>

<details>
<summary><strong><code>Integrated Services</code></strong></summary>
<br/>

- Unwriter's [bitpic](https://tncpw.co/e4d6ce84)
- Deggen's [Roundesk](https://tncpw.co/2d8d2e22) & [Baemail](https://tncpw.co/2c90c26b)
- RelayX's [Dime.ly](https://tncpw.co/46a4d32d)
- [2paymail](https://tncpw.co/c82bd419)
</details>

<details>
<summary><strong><code>Custom Configuration</code></strong></summary>
<br/>

The configuration file should be located in your `$HOME/paymail` folder and named `config.yaml`.

View the [example config file](config-example.yaml).

You can also specify a custom configuration file using `--config "/folder/path/file.yaml"`
</details>

<details>
<summary><strong><code>Local Database (Cache)</code></strong></summary>
<br/>

The database is located in your `$HOME/paymail` folder.

To clear the entire database:
```shell script
paymail --flush-cache
```

Run commands _ignoring_ local cache:
```shell script
paymail whois mrz --no-cache
```
</details>

<details>
<summary><strong><code>Package Dependencies</code></strong></summary>
<br/>

- [badger](https://github.com/dgraph-io/badger/v2) for persistent database storage
- [bsvd](https://github.com/bitcoinsv/bsvd) and [bsvutil](https://github.com/bitcoinsv/bsvutil) for BSV script functionality
- [chalk](https://github.com/ttacon/chalk) for colorful logs
- [cobra](https://github.com/spf13/cobra) and [viper](https://github.com/spf13/viper) for an easy configuration & CLI application development
- [columnize](https://github.com/ryanuber/columnize) for displaying terminal data in columns
- [dns](https://github.com/miekg/dns) package for advanced DNS functionality
- [go-homedir](https://github.com/mitchellh/go-homedir) to find the home directory
- [go-sanitize](https://github.com/mrz1836/go-sanitize) for sanitation and data formatting
- [go-validate](https://github.com/mrz1836/go-validate) for domain/email/ip validations
- [resty](https://github.com/go-resty/resty) for custom HTTP client support
</details>

<details>
<summary><strong><code>Application Deployment</code></strong></summary>
<br/>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary deployment to Github and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.

The release can also be deployed to a `homebrew` repository: [homebrew-paymail-inspector](https://github.com/mrz1836/homebrew-paymail-inspector).
</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>
<br/>

View all `makefile` commands
```shell script
make help
```

List of all current commands:
```text
all                    Runs multiple commands
clean                  Remove previous builds and any test cache data
clean-mods             Remove all the Go mod cache
coverage               Shows the test coverage
darwin                 Build for Darwin (macOS amd64)
gen-docs               Generate documentation from all available commands (fresh install)
gif-render             Render gifs in .github dir (find/replace text etc)
godocs                 Sync the latest tag with GoDocs
help                   Show this help message
install                Install the application
install-go             Install the application (Using Native Go)
lint                   Run the Go lint application
linux                  Build for Linux (amd64)
release                Full production release (creates release in Github)
release                Runs common.release then runs godocs
release-snap           Test the full release (build binaries)
release-test           Full production test release (everything except deploy)
replace-version        Replaces the version in HTML/JS (pre-deploy)
tag                    Generate a new tag and push (tag version=0.0.0)
tag-remove             Remove a tag if found (tag-remove version=0.0.0)
tag-update             Update an existing tag to current commit (tag-update version=0.0.0)
test                   Runs vet, lint and ALL tests
test-short             Runs vet, lint and tests (excludes integration tests)
test-travis            Runs all tests via Travis (also exports coverage)
test-travis-short      Runs unit tests via Travis (also exports coverage)
uninstall              Uninstall the application (and remove files)
vet                    Run the Go vet application
windows                Build for Windows (amd64)
```
</details>

<br/>

## Examples & Tests
All unit tests and [examples](docs/examples.md) run via [Travis CI](https://travis-ci.com/mrz1836/paymail-inspector) and uses [Go version 1.14.x](https://golang.org/doc/go1.14). View the [deployment configuration file](.travis.yml).

Run all tests (including integration tests)
```shell script
make test
```

<br/>

## Code Standards
Read more about this Go project's [code standards](CODE_STANDARDS.md).

<br/>

## Usage
View all the [examples](docs/examples.md) and see the [commands above](#commands)

All the generated command documentation can be found in [docs/commands](docs/commands).

<br/>

## Maintainers
| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) | [<img src="https://github.com/rohenaz.png" height="50" alt="Satchmo" />](https://github.com/rohenaz) |
|:---:|:---:|
| [MrZ](https://github.com/mrz1836) | [Satchmo](https://github.com/rohenaz) |

<br/>

## Contributing
View the [contributing guidelines](CONTRIBUTING.md) and please follow the [code of conduct](CODE_OF_CONDUCT.md).

### How can I help?
All kinds of contributions are welcome :raised_hands:! 
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:. 
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/mrz1836) :clap: 
or by making a [**bitcoin donation**](https://mrz1818.com/?tab=tips&af=paymail-inspector) to ensure this journey continues indefinitely! :rocket:

Help by sharing: [![Twetch](https://img.shields.io/badge/share-twitter-00ACEE.svg)](https://twitter.com/intent/tweet?text=Paymail%20Inspector%20Rocks!%20Check%20it%20out:%20https%3A%2F%2Ftncpw.co%2F2d429aee) [![Twitter](https://img.shields.io/badge/share-twetch-085AF6.svg)](https://twetch.app/compose?description=Paymail%20Inspector%20Rocks!%20Check%20it%20out:%20https%3A%2F%2Ftncpw.co%2F2d429aee)

### Credits
Inspiration and code snippets from [dnssec](https://github.com/binaryfigments/dnssec) and [check-ssl](https://github.com/wycore/check-ssl)

Utilized [terminalizer](https://terminalizer.com/) to record example gifs

<br/>

## License

![License](https://img.shields.io/github/license/mrz1836/paymail-inspector.svg?style=flat)
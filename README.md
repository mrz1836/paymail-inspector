<img src=".github/IMAGES/paymail-inspector.png" height="250" alt="Paymail Inspector">

# Paymail Inspector

[![Go](https://img.shields.io/github/go-mod/go-version/mrz1836/paymail-inspector)](https://golang.org/)
[![Build Status](https://travis-ci.com/mrz1836/paymail-inspector.svg?branch=master&v=2)](https://travis-ci.com/mrz1836/paymail-inspector)
[![Report](https://goreportcard.com/badge/github.com/mrz1836/paymail-inspector?style=flat&v=2)](https://goreportcard.com/report/github.com/mrz1836/paymail-inspector)
[![Release](https://img.shields.io/github/release-pre/mrz1836/paymail-inspector.svg?style=flat&v=2)](https://github.com/mrz1836/paymail-inspector/releases)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat)](https://github.com/RichardLitt/standard-readme)
[![GoDoc](https://godoc.org/github.com/mrz1836/paymail-inspector?status.svg&style=flat)](https://pkg.go.dev/github.com/mrz1836/paymail-inspector?tab=subdirectories)

> **paymail-inspector** is a CLI tool for inspecting, validating or resolving paymail addresses and domains

<img src=".github/IMAGES/paymail-inspector.gif?raw=true&v=2" alt="Paymail Commands">

## Table of Contents
- [Installation](#installation)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Benchmarks](#benchmarks)
- [Code Standards](#code-standards)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

## Installation

Run via [Go](https://formulae.brew.sh/formula/go)
```bash
$ go get -u github.com/mrz1836/paymail-inspector && go install github.com/mrz1836/paymail-inspector
$ paymail-inspector -h
```


Run via [a binary](https://github.com/mrz1836/paymail-inspector/releases) on Linux, Mac or Windows (Mac example)
```bash
$ curl -LkSs https://github.com/mrz1836/paymail-inspector/releases/download/v0.0.16/paymail-inspector_macOS_64-bit.tar.gz -o app.tar.gz
$ tar -zxf app.tar.gz
$ ./app/paymail-inspector -h
```

## Commands

### `brfc`
List all known brfc specifications ([view example](docs/examples.md#list-brfc-specifications))
```bash
$ paymail-inspector brfc list
```

Generate a new `BRFC ID` for a new specification ([view example](docs/examples.md#generate-new-brfc-id))
```bash
$ paymail-inspector brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```

### `capabilities`
Lists the available capabilities of the paymail service ([view example](docs/examples.md#get-capabilities-by-domain))
```bash
$ paymail-inspector capabilities simply.cash
```

### `p2p`
Starts a p2p payment request and returns (n) outputs of (`script`,`satoshis`,`address`) ([view example](docs/examples.md#start-p2p-payment-request-by-paymail))
```bash
$ paymail-inspector p2p mrz@handcash.io
```

### `resolve`
Returns the `pubkey`, `output script` and `address` for a given paymail address ([view example](docs/examples.md#resolve-paymail-address-by-paymail))
```bash
$ paymail-inspector resolve mrz@simply.cash
```

### `validate`
Runs several validations on the paymail service for DNSSEC, SSL, SRV and required capabilities ([view example](docs/examples.md#validate-paymail-setup-by-paymail-or-domain))
```bash
$ paymail-inspector validate simply.cash --skip-dnssec
```

### `verify`
Verifies if a paymail is associated to a pubkey ([view example](docs/examples.md#verify-public-key-owner))
```bash
$ paymail-inspector verify mrz@simply.cash 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec
```

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
- [x] Public Profile (f12f968c92d6) (unknown source)
- [x] P2P Payment Destination ([2a40af698840](https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html))
- [ ] P2P Transactions ([5f1323cddf31](https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html))
- [ ] Sender Validation ([6745385c3fc0](http://bsvalias.org/04-02-sender-validation.html))
- [ ] Receiver Approvals ([3d7c2ca83a46](http://bsvalias.org/04-03-receiver-approvals.html))

### Package Dependencies
- bitcoinsv's [bsvd](https://github.com/bitcoinsv/bsvd) for BSV script functionality
- bitcoinsv's [bsvutil](https://github.com/bitcoinsv/bsvutil) for BSV address utilities
- miekg's [dns](https://github.com/miekg/dns) package for advanced DNS functionality
- mitchellh's [go-homedir](https://github.com/mitchellh/go-homedir) to find the home directory
- MrZ's [go-validate](https://github.com/mrz1836/go-validate) for domain/email/ip validations
- spf13's [cobra](https://github.com/spf13/cobra) for easy CLI application development
- spf13's [viper](https://github.com/spf13/viper) for easy application configuration
- ttacon's [chalk](https://github.com/ttacon/chalk) for colorful logs

### Optional Add-ons
- [goreleaser](https://github.com/goreleaser/goreleaser) for easy binary deployment to Github (`brew install goreleaser`)

#### Upgrade Dependencies & Reinstall
```bash
$ make update
$ make install
```

#### Uninstall Application
```bash
$ make uninstall
```

#### View all `makefile` commands
```bash
$ make help
```

#### Deploy a New Version
```bash
$ export GITHUB_TOKEN=<your_token>
$ make tag version=0.1.0
$ make release
```

#### Custom Configuration
The file should be located in your `$HOME` folder and named `.paymail-inspector.yaml`. View the [example config file](.paymail-inspector.yaml).

## Examples & Tests
All unit tests and [examples](docs/examples.md) run via [Travis CI](https://travis-ci.com/mrz1836/paymail-inspector) and uses [Go version 1.14.x](https://golang.org/doc/go1.14). View the [deployment configuration file](.travis.yml).

Run all tests (including integration tests)
```bash
$ make test
```

## Code Standards
Read more about this Go project's [code standards](CODE_STANDARDS.md).

## Usage
View all the [examples](docs/examples.md) and see the [commands above](#commands)

All the generated command documentation can be found in [docs/commands](docs/commands).

## Maintainers

| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) | [<img src="https://github.com/rohenaz.png" height="50" alt="Satchmo" />](https://github.com/rohenaz) |
|:---:|:---:|
| [MrZ](https://github.com/mrz1836) | [Satchmo](https://github.com/rohenaz) |


## Contributing

Inspiration and code snippets from [dnssec](https://github.com/binaryfigments/dnssec) and [check-ssl](https://github.com/wycore/check-ssl)

Utilized [terminalizer](https://terminalizer.com/) to record example gifs

View the [contributing guidelines](CONTRIBUTING.md) and follow the [code of conduct](CODE_OF_CONDUCT.md).

Support the development of this project 🙏

[![Donate](https://img.shields.io/badge/donate-bitcoin-brightgreen.svg)](https://mrz1818.com/?tab=tips&af=paymail-inspector)

## License

![License](https://img.shields.io/github/license/mrz1836/paymail-inspector.svg?style=flat)
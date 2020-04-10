<img src=".github/IMAGES/paymail-inspector.png" height="250" alt="Paymail Inspector">

**paymail-inspector** is a CLI tool for inspecting, validating and resolving paymail addresses and domains

[![Go](https://img.shields.io/github/go-mod/go-version/mrz1836/paymail-inspector?v=1)](https://golang.org/)
[![Build Status](https://travis-ci.com/mrz1836/paymail-inspector.svg?branch=master&v=1)](https://travis-ci.com/mrz1836/paymail-inspector)
[![Report](https://goreportcard.com/badge/github.com/mrz1836/paymail-inspector?style=flat&v=1)](https://goreportcard.com/report/github.com/mrz1836/paymail-inspector)
[![Release](https://img.shields.io/github/release-pre/mrz1836/paymail-inspector.svg?style=flat&v=1)](https://github.com/mrz1836/paymail-inspector/releases)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat)](https://github.com/RichardLitt/standard-readme)
[![GoDoc](https://godoc.org/github.com/mrz1836/paymail-inspector?status.svg&style=flat)](https://godoc.org/github.com/mrz1836/paymail-inspector)

<img src=".github/IMAGES/capabilities-command-zoomed.gif?raw=true" alt="Capabilities Command">

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

**paymail-inspector** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).
```bash
$ go get -u github.com/mrz1836/paymail-inspector
$ go install github.com/mrz1836/paymail-inspector
```

#### Run the Application
```bash
$ paymail-inspector -h
```

### Package Dependencies
- bitcoinsv's [bsvd](https://github.com/bitcoinsv/bsvd) for BSV script functionality
- bitcoinsv's [bsvutil](https://github.com/bitcoinsv/bsvutil) for BSV address utilities
- miekg's [dns](https://github.com/miekg/dns) package for advanced DNS functionality
- mitchellh's [go-homedir](https://github.com/mitchellh/go-homedir) to find the home directory
- MrZ's [go-validate](https://github.com/mrz1836/go-validate) for domain/email/ip validations
- spf13's [cobra](https://github.com/spf13/cobra) for easy CLI application development
- spf13's [viper](https://github.com/spf13/viper) for easy application configuration

#### Upgrade Dependencies
```bash
$ go get -u ./...
$ go mod tidy
```

## Documentation
You can view the generated [documentation here](https://godoc.org/github.com/mrz1836/paymail-inspector).

Also checkout the [official paymail specifications](http://bsvalias.org/index.html).
Additional information can also be found via [MoneyButton's documentation](https://docs.moneybutton.com/docs/paymail-overview.html).

### Features
- [x] Validate any paymail domain or paymail address
- [x] Customize the validation requirements via application flags
- [x] Validate the SRV record, DNSSEC and SSL for the target domain
- [x] Validation for required capabilities (`pki`, `paymentDestination`)
- [x] List paymail capabilities ([.well-known/bsvalias](http://bsvalias.org/02-02-capability-discovery.html))
- [x] Validate the `pki` response (brfc: [0c4339ef99c2](http://bsvalias.org/03-public-key-infrastructure.html))
- [x] Resolve a paymail address (brfc: [759684b1a19a](http://bsvalias.org/04-01-basic-address-resolution.html))
- [ ] Sender validation (brfc: [6745385c3fc0](http://bsvalias.org/04-02-sender-validation.html))
- [ ] Receiver approvals (brfc: [3d7c2ca83a46](http://bsvalias.org/04-03-receiver-approvals.html))
- [x] PayTo protocol prefix (brfc: [7bd25e5a1fc6](http://bsvalias.org/04-04-payto-protocol-prefix.html))
- [ ] Verify public key owner (brfc: [a9f510c16bde](http://bsvalias.org/05-verify-public-key-owner.html))
- [ ] P2P Transactions (brfc: [5f1323cddf31](https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html))
- [ ] P2P Payment Destination (brfc: [2a40af698840](https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html))

## Examples & Tests
All unit tests and [examples](examples/examples.md) run via [Travis CI](https://travis-ci.com/mrz1836/paymail-inspector) and uses [Go version 1.14.x](https://golang.org/doc/go1.14). View the [deployment configuration file](.travis.yml).

Run all tests (including integration tests)
```bash
$ cd ../paymail-inspector
$ go test ./... -v
```

Run tests (excluding integration tests)
```bash
$ cd ../paymail-inspector
$ go test ./... -v -test.short
```

## Code Standards
Read more about this Go project's [code standards](CODE_STANDARDS.md).

## Usage
View the [examples](examples/examples.md)

## Maintainers

| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) |
|:---:|
| [MrZ](https://github.com/mrz1836) |


## Contributing

Inspiration and code snippets from [dnssec](https://github.com/binaryfigments/dnssec) and [check-ssl](https://github.com/wycore/check-ssl)

Utilized [terminalizer](https://terminalizer.com/) to record example gifs

View the [contributing guidelines](CONTRIBUTING.md) and follow the [code of conduct](CODE_OF_CONDUCT.md).

Support the development of this project 🙏

[![Donate](https://img.shields.io/badge/donate-bitcoin-brightgreen.svg)](https://mrz1818.com/?tab=tips&af=paymail-inspector)

## License

![License](https://img.shields.io/github/license/mrz1836/paymail-inspector.svg?style=flat)
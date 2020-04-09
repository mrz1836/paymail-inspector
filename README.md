# paymail-inspector
**paymail-inspector** is a CLI tool for inspecting paymail addresses and domains

[![Go](https://img.shields.io/github/go-mod/go-version/mrz1836/paymail-inspector)](https://golang.org/)
[![Build Status](https://travis-ci.com/mrz1836/paymail-inspector.svg?branch=master)](https://travis-ci.com/mrz1836/paymail-inspector)
[![Report](https://goreportcard.com/badge/github.com/mrz1836/paymail-inspector?style=flat)](https://goreportcard.com/report/github.com/mrz1836/paymail-inspector)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/0b377a0d1dde4b6ba189545aa7ee2e17)](https://www.codacy.com/app/mrz1818/paymail-inspector?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=mrz1836/paymail-inspector&amp;utm_campaign=Badge_Grade)
[![Release](https://img.shields.io/github/release-pre/mrz1836/paymail-inspector.svg?style=flat)](https://github.com/mrz1836/paymail-inspector/releases)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat)](https://github.com/RichardLitt/standard-readme)
[![GoDoc](https://godoc.org/github.com/mrz1836/paymail-inspector?status.svg&style=flat)](https://godoc.org/github.com/mrz1836/paymail-inspector)

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
```

#### Run the Application
```bash
$ paymail-inspector -h
```

#### Reinstall (after making modifications)
```bash
$ go install github.com/mrz1836/paymail-inspector
```

#### Upgrade Dependencies
```bash
$ go get -u ./...
$ go mod tidy
```

### Package Dependencies
- miekg's [dns](https://github.com/miekg/dns) package for advanced DNS functionality
- mitchellh's [go-homedir](https://github.com/mitchellh/go-homedir) to find the home directory
- MrZ's [go-validate](https://github.com/mrz1836/go-validate) for domain/email/ip validations
- spf13's [cobra](https://github.com/spf13/cobra) for easy CLI application development
- spf13's [viper](https://github.com/spf13/viper) for easy application configuration

## Documentation
You can view the generated [documentation here](https://godoc.org/github.com/mrz1836/paymail-inspector).

### Features
- [x] Validate any domain or paymail address
- [x] Check the SRV record, DNSSEC and SSL for the target domain
- [x] Check for required capabilities (pki, paymentDestination)
- [x] Validate the pki response
- [x] List paymail capabilities
- [ ] Resolve a paymail address

## Examples & Tests
All unit tests and [examples](examples/examples.go) run via [Travis CI](https://travis-ci.com/mrz1836/paymail-inspector) and uses [Go version 1.14.x](https://golang.org/doc/go1.14). View the [deployment configuration file](.travis.yml).

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
View the [examples](examples/examples.go)

## Maintainers

| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) |
|:---:|
| [MrZ](https://github.com/mrz1836) |


## Contributing

Inspiration and code snippets from [dnssec](https://github.com/binaryfigments/dnssec) and [check-ssl](https://github.com/wycore/check-ssl)

Utilized [terminalizer](https://terminalizer.com/) to record cool gifs!

View the [contributing guidelines](CONTRIBUTING.md) and follow the [code of conduct](CODE_OF_CONDUCT.md).

Support the development of this project 🙏

[![Donate](https://img.shields.io/badge/donate-bitcoin-brightgreen.svg)](https://mrz1818.com/?tab=tips&af=paymail-inspector)

## License

![License](https://img.shields.io/github/license/mrz1836/paymail-inspector.svg?style=flat)
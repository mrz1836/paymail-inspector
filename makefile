COVER=go tool cover

## Default Repo Domain
GIT_DOMAIN=github.com

## Automatically detect the repo owner and repo name
REPO_NAME=$(shell basename `git rev-parse --show-toplevel`)
REPO_OWNER=$(shell git config --get remote.origin.url | sed 's/git@$(GIT_DOMAIN)://g' | sed 's/\/$(REPO_NAME).git//g')

## Symlink into GOPATH
BUILD_DIR=${GOPATH}/src/${GIT_DOMAIN}/${REPO_OWNER}/${REPO_NAME}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})
RELEASES_DIR=./releases

## Set the binary release names
DARWIN=$(REPO_NAME)-darwin
LINUX=$(REPO_NAME)-linux
WINDOWS=$(REPO_NAME)-windows.exe

## Set the version(s) (injected into binary)
VERSION=$(shell git describe --tags --always --long --dirty)
VERSION_SHORT=$(shell git describe --tags --always --abbrev=0)

.PHONY: test install clean release link

all: test install gen-docs ## Runs test, install, clean, docs

bench:  ## Run all benchmarks in the Go application
	go test -bench ./... -benchmem -v

build-go:  ## Build the Go application (locally)
	go build -o bin/$(REPO_NAME)

build: darwin linux windows ## Build all binaries (darwin, linux, windows)
	@echo version: $(VERSION_SHORT)

clean: ## Remove previous builds and any test cache data
	go clean -cache -testcache -i -r
	if [ -d ${RELEASES_DIR} ]; then rm -r ${RELEASES_DIR}; fi

clean-mods: ## Remove all the Go mod cache
	go clean -modcache

coverage: ## Shows the test coverage
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

darwin: $(DARWIN) ## Build for Darwin (macOS amd64)

gen-docs: ## Generate documentation from all available commands (fresh install)
	make install
	$(REPO_NAME) --docs

godocs: ## Sync the latest tag with GoDocs
	curl https://proxy.golang.org/$(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/@v/$(VERSION_SHORT).info

help: ## Show all make commands available
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install: ## Install the application
	go install $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)

link:
	BUILD_DIR=${BUILD_DIR}; \
	BUILD_DIR_LINK=${BUILD_DIR_LINK}; \
	CURRENT_DIR=${CURRENT_DIR}; \
	if [ "$${BUILD_DIR_LINK}" != "$${CURRENT_DIR}" ]; then \
	    echo "Fixing symlinks for build"; \
	    rm -f $${BUILD_DIR}; \
	    ln -s $${CURRENT_DIR} $${BUILD_DIR}; \
	fi

lint: ## Run the Go lint application
	golint

linux: $(LINUX) ## Build for Linux (amd64)

release: ## Full production release (creates release in Github)
	 goreleaser

release-test: ## Full production test release (everything except deploy)
	 goreleaser --skip-publish

release-snap: ## Test the full release (build binaries)
	goreleaser --snapshot --skip-publish --rm-dist

run: ## Runs the go application
	go run main.go

test: ## Runs vet, lint and ALL tests
	go vet -v
	golint
	go test ./... -v

test-short: ## Runs vet, lint and tests (excludes integration tests)
	go vet -v
	golint
	go test ./... -v -test.short

uninstall: ## Uninstall the application (and remove files)
	go clean -i $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)
	rm -rf $$GOPATH/src/$(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)

update:  ## Update all project dependencies
	go get -u ./...
	go mod tidy

vet: ## Run the Go vet application
	go vet -v

windows: $(WINDOWS) ## Build for Windows (amd64)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -i -v -o ${RELEASES_DIR}/$(WINDOWS) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o ${RELEASES_DIR}/$(LINUX) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o ${RELEASES_DIR}/$(DARWIN) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"
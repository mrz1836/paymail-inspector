COVER=go tool cover

## Default Repo Domain
GIT_DOMAIN=github.com

## Custom application binary name
CUSTOM_BINARY_NAME=paymail

## Set the Github Token
#GITHUB_TOKEN=<your_token>

## Automatically detect the repo owner and repo name
REPO_NAME=$(shell basename `git rev-parse --show-toplevel`)
REPO_OWNER=$(shell git config --get remote.origin.url | sed 's/git@$(GIT_DOMAIN)://g' | sed 's/\/$(REPO_NAME).git//g')

ifeq ($(CUSTOM_BINARY_NAME),)
CUSTOM_BINARY_NAME := $(REPO_NAME)
endif

## Symlink into GOPATH
BUILD_DIR=${GOPATH}/src/${GIT_DOMAIN}/${REPO_OWNER}/${REPO_NAME}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})
DISTRIBUTIONS_DIR=./dist

## Set the binary release names
DARWIN=$(CUSTOM_BINARY_NAME)-darwin
LINUX=$(CUSTOM_BINARY_NAME)-linux
WINDOWS=$(CUSTOM_BINARY_NAME)-windows.exe

## Set the version(s) (injected into binary)
VERSION=$(shell git describe --tags --always --long --dirty)
VERSION_SHORT=$(shell git describe --tags --always --abbrev=0)

.PHONY: test install clean release link

all: test install gen-docs ## Runs test, install, clean, docs

bench:  ## Run all benchmarks in the Go application
	go test -bench ./... -benchmem -v

build-go:  ## Build the Go application (locally)
	go build -o bin/$(CUSTOM_BINARY_NAME)

build: darwin linux windows ## Build all binaries (darwin, linux, windows)
	@echo version: $(VERSION_SHORT)

clean: ## Remove previous builds and any test cache data
	go clean -cache -testcache -i -r
	if [ -d ${DISTRIBUTIONS_DIR} ]; then rm -r ${DISTRIBUTIONS_DIR}; fi

clean-mods: ## Remove all the Go mod cache
	go clean -modcache

coverage: ## Shows the test coverage
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

darwin: $(DARWIN) ## Build for Darwin (macOS amd64)

gen-docs: ## Generate documentation from all available commands (fresh install)
	make install
	$(CUSTOM_BINARY_NAME) --docs

gif-render: ## Render gifs in .github dir (find/replace text etc)
	test $(name)
	find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/Terminalizer/null/g"
	find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/MrZs-MacBook-Pro:paymail-inspector MrZ//g"
	find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/floating/solid/g"
	find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/logout//g"
	terminalizer render gif -o $(name)
	cp -f *.gif .github/IMAGES/
	rm -rf *.gif && rm -rf gif.yml

godocs: ## Sync the latest tag with GoDocs
	curl https://proxy.golang.org/$(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/@v/$(VERSION_SHORT).info

help: ## Show all make commands available
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install: ## Install the application
	go build -o $$GOPATH/bin/$(CUSTOM_BINARY_NAME)

install-go: ## Install the application (Using Native Go)
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
	 goreleaser --rm-dist
	 make godocs

release-test: ## Full production test release (everything except deploy)
	 goreleaser --skip-publish --rm-dist

release-snap: ## Test the full release (build binaries)
	goreleaser --snapshot --skip-publish --rm-dist

run: ## Runs the go application
	go run main.go

tag: ## Generate a new tag and push (IE: make tag version=0.0.0)
	test $(version)
	git tag -a v$(version) -m "Pending full release..."
	git push origin v$(version)
	git fetch --tags -f

tag-remove: ## Remove a tag if found (IE: make tag-remove version=0.0.0)
	test $(version)
	git tag -d v$(version)
	git push --delete origin v$(version)
	git fetch --tags

tag-update: ## Update an existing tag to current commit (IE: make tag-update version=0.0.0)
	test $(version)
	git push --force origin HEAD:refs/tags/v$(version)
	git fetch --tags -f

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
	rm -rf $$GOPATH/bin/$(CUSTOM_BINARY_NAME)

update:  ## Update all project dependencies
	go get -u ./...
	go mod tidy

update-releaser:  ## Update the goreleaser application
	brew update
	brew upgrade goreleaser

vet: ## Run the Go vet application
	go vet -v

windows: $(WINDOWS) ## Build for Windows (amd64)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(WINDOWS) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(LINUX) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(DARWIN) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"
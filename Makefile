## Custom binary
CUSTOM_BINARY_NAME := paymail

# Common makefile commands & variables between projects
include .make/Makefile.common

# Common Golang makefile commands & variables between projects
include .make/Makefile.go

## Not defined? Use default repo name which is the application
ifeq ($(REPO_NAME),)
	REPO_NAME="paymail-inspector"
endif

## Not defined? Use default repo owner
ifeq ($(REPO_OWNER),)
	REPO_OWNER="mrz1836"
endif

.PHONY: build clean release

all: ## Runs multiple commands
	@$(MAKE) test
	@$(MAKE) gen-docs

build:  ## Build all binaries (darwin, linux, windows)
	@$(MAKE) darwin
	@$(MAKE) linux
	@$(MAKE) windows
	@echo version: $(VERSION_SHORT)

clean: ## Remove previous builds and any test cache data
	@go clean -cache -testcache -i -r
	@if [ -d $(DISTRIBUTIONS_DIR) ]; then rm -r $(DISTRIBUTIONS_DIR); fi

darwin: ## Build for Darwin (macOS amd64)
	@$(MAKE) $(DARWIN)

gen-docs: ## Generate documentation from all available commands (fresh install)
	@$(MAKE) install
	@$(CUSTOM_BINARY_NAME) --docs

gif-render: ## Render gifs in .github dir (find/replace text etc)
	@test $(name)
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/Terminalizer/null/g"
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/MrZs-MacBook-Pro:paymail-inspector MrZ//g"
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/floating/solid/g"
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/logout//g"
	@terminalizer render gif -o $(name)
	@cp -f *.gif .github/IMAGES/
	@rm -rf *.gif && rm -rf gif.yml

linux: ## Build for Linux (amd64)
	@$(MAKE) $(LINUX)

release:: ## Runs common.release then runs godocs
	@$(MAKE) godocs

update-terminalizer:  ## Update the terminalizer application
	@npm update -g terminalizer

windows: ## Build for Windows (amd64)
	@$(MAKE) $(WINDOWS)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(WINDOWS) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(LINUX) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(DARWIN) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"
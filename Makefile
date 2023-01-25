## Custom binary
CUSTOM_BINARY_NAME := paymail

# Common makefile commands & variables between projects
include .make/common.mk

# Common Golang makefile commands & variables between projects
include .make/go.mk

## Not defined? Use default repo name which is the application
ifeq ($(REPO_NAME),)
	REPO_NAME="paymail-inspector"
endif

## Not defined? Use default repo owner
ifeq ($(REPO_OWNER),)
	REPO_OWNER="mrz1836"
endif

.PHONY: all
all: ## Runs multiple commands
	@$(MAKE) test
	@$(MAKE) gen-docs

.PHONY: build
build: ## Build all binaries (darwin, linux, windows)
	@$(MAKE) darwin
	@$(MAKE) linux
	@$(MAKE) windows
	@echo version: $(VERSION_SHORT)

.PHONY: clean
clean: ## Remove previous builds and any test cache data
	@go clean -cache -testcache -i -r
	@test $(DISTRIBUTIONS_DIR)
	@if [ -d $(DISTRIBUTIONS_DIR) ]; then rm -r $(DISTRIBUTIONS_DIR); fi

.PHONY: darwin
darwin: ## Build for Darwin (macOS amd64)
	@$(MAKE) $(DARWIN)

.PHONY: gen-docs
gen-docs: ## Generate documentation from all available commands (fresh install)
	@$(MAKE) install
	@$(CUSTOM_BINARY_NAME) --docs

.PHONY: gif-render
gif-render: ## Render gifs in .github dir (find/replace text etc)
	@test $(name)
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/Terminalizer/null/g"
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/MrZs-MacBook-Pro:paymail-inspector MrZ//g"
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/floating/solid/g"
	@find . -name 'gif.yml' -print0 | xargs -0 sed -i "" "s/logout//g"
	@terminalizer render gif -o $(name)
	@cp -f *.gif .github/IMAGES/
	@rm -rf *.gif && rm -rf gif.yml

.PHONY: linux
linux: ## Build for Linux (amd64)
	@$(MAKE) $(LINUX)

.PHONY: release
release:: ## Runs common.release then runs godocs
	@$(MAKE) godocs

.PHONY: update-terminalizer
update-terminalizer: ## Update the terminalizer application
	@npm update -g terminalizer

.PHONY: windows
windows: ## Build for Windows (amd64)
	@$(MAKE) $(WINDOWS)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(WINDOWS) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(LINUX) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o ${DISTRIBUTIONS_DIR}/$(DARWIN) -ldflags="-s -w -X $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/cmd.Version=$(VERSION_SHORT)"
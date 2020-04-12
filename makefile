COVER=go tool cover

test:
	go vet
	golint
	go test ./... -v

install:
	go install github.com/mrz1836/paymail-inspector

uninstall:
	go clean -i github.com/mrz1836/paymail-inspector
	rm -rf $$GOPATH/src/github.com/mrz1836/paymail-inspector

build:
	go build -o bin/paymail-inspector

update:
	go get -u ./...
	go mod tidy

release:
	chmod +x scripts/release
	scripts/release build
	curl https://proxy.golang.org/github.com/mrz1836/paymail-inspector/@v/v0.0.12.info

clean:
	go clean -testcache
	$(RM) -r release bin

all: test build release

.PHONY: install test clean release
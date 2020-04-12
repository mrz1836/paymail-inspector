COVER=go tool cover

test:
	go vet
	golint
	go test ./... -v

test-short:
	go vet
	golint
	go test ./... -v -test.short

install:
	go install github.com/mrz1836/paymail-inspector

uninstall:
	go clean -i github.com/mrz1836/paymail-inspector
	rm -rf $$GOPATH/src/github.com/mrz1836/paymail-inspector

build:
	go build -o bin/paymail-inspector

gen-docs:
	paymail-inspector --docs
	
update:
	go get -u ./...
	go mod tidy

release:
	curl https://proxy.golang.org/github.com/mrz1836/paymail-inspector/@v/v0.0.14.info

clean:
	go clean -testcache

all: test install clean gen-docs release

.PHONY: test install clean gen-docs release
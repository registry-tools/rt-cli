.ONESHELL:
.SHELLFLAGS += -e

default: build

build:
	go build -ldflags "-X main.version=dev" -o bin/rt cmd/rt/main.go

test:
	go test ./... -v

.PHONY: test build

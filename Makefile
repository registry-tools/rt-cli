.ONESHELL:
.SHELLFLAGS += -e

default: build

build:
	go build -ldflags "-X main.version=dev" -o bin/rt cmd/rt/main.go

lint:
	docker run --rm -v ./:/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run -v

test:
	go test ./... -v

.PHONY: lint test build

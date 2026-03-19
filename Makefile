BIN := slack-cli
OUT := out
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build install clean test vet

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(OUT)/$(BIN) ./cmd/slack-cli

install:
	go install -ldflags "-X main.version=$(VERSION)" ./cmd/slack-cli

clean:
	rm -rf $(OUT)

test:
	go test ./... -count=1

vet:
	go vet ./...

PROJECT = rain
MAIN_FILE = ./cmd/$(PROJECT)/main.go
INSTALL_PREFIX = ${HOME}/.local/bin/

lint:
	@golangci-lint run ./...

test:
	@go test ./...

dev:
	@go run $(MAIN_FILE)

build:
	go build -o ./bin/$(PROJECT) $(MAIN_FILE)

install: build
	mv -v ./bin/$(PROJECT) $(INSTALL_PREFIX)

.PHONY: lint test dev

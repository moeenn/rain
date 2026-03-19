PROJECT = rain
MAIN_FILE = ./cmd/$(PROJECT)/main.go

lint:
	@golangci-lint run ./...

test:
	@go test ./...

dev:
	@go run $(MAIN_FILE)

build: lint test
	go build -o ./bin/$(PROJECT) $(MAIN_FILE)

.PHONY: lint test dev
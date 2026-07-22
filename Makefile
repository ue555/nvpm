.PHONY: build run install clean test fmt

# Build the application
build:
	go build -o bin/nvpm ./cmd/nvpm

# Run the application
run:
	go run ./cmd/nvpm

# Install dependencies
install:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f nvpm-lock.json

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Run with example config
example:
	go run ./cmd/nvpm -config examples/config.json -cmd stats

# Execute commands
cmd-install:
	go run ./cmd/nvpm -config examples/config.json -cmd install

cmd-update:
	go run ./cmd/nvpm -config examples/config.json -cmd update

cmd-sync:
	go run ./cmd/nvpm -config examples/config.json -cmd sync

cmd-list:
	go run ./cmd/nvpm -config examples/config.json -cmd list

cmd-stats:
	go run ./cmd/nvpm -config examples/config.json -cmd stats

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  install     - Install dependencies"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  example     - Run with example config"
	@echo "  cmd-install - Install plugins"
	@echo "  cmd-update  - Update plugins"
	@echo "  cmd-sync    - Sync plugins"
	@echo "  cmd-list    - List plugins"
	@echo "  cmd-stats   - Show statistics"

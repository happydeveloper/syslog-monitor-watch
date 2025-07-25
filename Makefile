BINARY_NAME=syslog-monitor
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_LINUX=$(BINARY_NAME)_linux

.PHONY: all build clean test install

all: clean build

build:
	go mod tidy
	go build -o $(BINARY_NAME) -v .

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_LINUX) .

build-unix:
	CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo -o $(BINARY_UNIX) .

install:
	go mod download
	go mod tidy

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_LINUX)

run:
	go run main.go

help:
	@echo "Available targets:"
	@echo "  build       - Build the binary for current platform"
	@echo "  build-linux - Build the binary for Linux"
	@echo "  build-unix  - Build the binary for macOS/Unix"
	@echo "  install     - Download dependencies"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run the application"
	@echo "  help        - Show this help message" 
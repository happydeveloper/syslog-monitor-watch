BINARY_NAME=syslog-monitor
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_LINUX=$(BINARY_NAME)_linux
BINARY_MACOS=$(BINARY_NAME)_macos
BINARY_MACOS_ARM=$(BINARY_NAME)_macos_arm64
BINARY_MACOS_INTEL=$(BINARY_NAME)_macos_amd64

.PHONY: all build clean test install build-all build-macos build-macos-universal install-macos

all: clean build

build:
	go mod tidy
	go build -o $(BINARY_NAME) -v .

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_LINUX) .

build-unix:
	CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo -o $(BINARY_UNIX) .

build-macos:
	@echo "Building for macOS (current architecture)..."
	CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo -o $(BINARY_MACOS) .

build-macos-arm64:
	@echo "Building for macOS Apple Silicon (ARM64)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -installsuffix cgo -o $(BINARY_MACOS_ARM) .

build-macos-intel:
	@echo "Building for macOS Intel (AMD64)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o $(BINARY_MACOS_INTEL) .

build-macos-universal:
	@echo "Building universal macOS binary..."
	$(MAKE) build-macos-arm64
	$(MAKE) build-macos-intel
	lipo -create -output $(BINARY_NAME)_macos_universal $(BINARY_MACOS_ARM) $(BINARY_MACOS_INTEL)
	@echo "Universal binary created: $(BINARY_NAME)_macos_universal"

build-all:
	@echo "Building for all platforms..."
	$(MAKE) build
	$(MAKE) build-linux
	$(MAKE) build-macos-universal

install:
	go mod download
	go mod tidy

test:
	go test -v ./...

install-macos:
	@echo "Installing to /usr/local/bin..."
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete! Run 'syslog-monitor -help' to get started."

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_LINUX)
	rm -f $(BINARY_MACOS)
	rm -f $(BINARY_MACOS_ARM)
	rm -f $(BINARY_MACOS_INTEL)
	rm -f $(BINARY_NAME)_macos_universal

run:
	go run main.go

help:
	@echo "🔧 Available targets:"
	@echo ""
	@echo "📦 Build targets:"
	@echo "  build                - Build the binary for current platform"
	@echo "  build-linux         - Build the binary for Linux"
	@echo "  build-unix          - Build the binary for macOS/Unix (legacy)"
	@echo "  build-macos         - Build for macOS (current architecture)"
	@echo "  build-macos-arm64   - Build for macOS Apple Silicon (ARM64)"
	@echo "  build-macos-intel   - Build for macOS Intel (AMD64)"
	@echo "  build-macos-universal - Build universal macOS binary (ARM64 + Intel)"
	@echo "  build-all           - Build for all platforms"
	@echo ""
	@echo "🛠️  Development:"
	@echo "  install      - Download dependencies"
	@echo "  install-macos - Install binary to /usr/local/bin (macOS)"
	@echo "  test         - Run tests"
	@echo "  run          - Run the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "🍎 macOS Quick Start:"
	@echo "  make build-macos && sudo make install-macos" 
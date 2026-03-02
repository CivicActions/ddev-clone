BINARY_NAME := ddev-clone
MODULE := github.com/civicactions/ddev-clone
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
BUILD_DIR := bin

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: build build-all test lint clean release fmt vet

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

build-all:
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Building $$os/$$arch..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) \
			-o $(BUILD_DIR)/$(BINARY_NAME)_$${os}_$${arch}$${ext} .; \
	done

test:
	go test -v -race ./pkg/... ./cmd/...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

release: clean build-all
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		tarball="dist/$(BINARY_NAME)_$${os}_$${arch}.tar.gz"; \
		echo "Packaging $$tarball..."; \
		tar -czf $$tarball -C $(BUILD_DIR) $(BINARY_NAME)_$${os}_$${arch}$${ext}; \
	done

coverage:
	go test -coverprofile=coverage.out ./pkg/... ./cmd/...
	go tool cover -html=coverage.out -o coverage.html

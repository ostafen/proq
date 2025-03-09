APP_NAME = proq
VERSION ?= $(shell git describe --tags --always --dirty=-dev)
BUILD_DIR = bin
CMD_DIR = cmd
ENTRY_POINT = $(CMD_DIR)/main.go

# Supported GOOS and GOARCH combinations
PLATFORMS = \
    linux/amd64 \
    linux/arm64 \
    linux/arm \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64

# Build for all platforms
.PHONY: build
build: clean $(PLATFORMS)

$(PLATFORMS):
	@mkdir -p $(BUILD_DIR)
	@GOOS=$(word 1, $(subst /, ,$@)) GOARCH=$(word 2, $(subst /, ,$@)) \
	    go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME)-$(word 1, $(subst /, ,$@))-$(word 2, $(subst /, ,$@)) $(ENTRY_POINT)
	@echo "Built $(APP_NAME) for $@ with version $(VERSION)"

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	@echo "Cleaned up build artifacts"

# Build for the current OS/Arch
.PHONY: build-local
build-local:
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) $(ENTRY_POINT)
	@echo "Built $(APP_NAME) for local machine with version $(VERSION)"
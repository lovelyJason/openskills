BINARY_NAME := openskills
SHORT_NAME := osk
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
LDFLAGS := -ldflags "-s -w -X github.com/lovelyJason/openskills/internal/cli.Version=$(VERSION)"
PREFIX := $(shell brew --prefix 2>/dev/null || echo /usr/local)
BUILD_DIR := $(CURDIR)
BIN_PATH := $(PREFIX)/bin/$(BINARY_NAME)
SHORT_BIN_PATH := $(PREFIX)/bin/$(SHORT_NAME)
BACKUP_PATH := $(PREFIX)/bin/.$(BINARY_NAME).brew-backup

.PHONY: build clean test install uninstall coverage fmt vet lint tidy snapshot

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/openskills

install: build
	@if [ -L "$(BIN_PATH)" ] && readlink "$(BIN_PATH)" | grep -q Cellar; then \
		cp -P "$(BIN_PATH)" "$(BACKUP_PATH)"; \
		echo "Backed up Homebrew link: $(BACKUP_PATH)"; \
	fi
	@ln -sf "$(BUILD_DIR)/$(BINARY_NAME)" "$(BIN_PATH)"
	@ln -sf "$(BIN_PATH)" "$(SHORT_BIN_PATH)"
	@echo "Linked: $(BIN_PATH) -> $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Linked: $(SHORT_BIN_PATH) -> $(BIN_PATH)"
	@echo "Run 'make uninstall' to restore Homebrew version"

uninstall:
	@rm -f "$(SHORT_BIN_PATH)" 2>/dev/null && echo "Removed: $(SHORT_BIN_PATH)" || true
	@if [ -L "$(BIN_PATH)" ] && readlink "$(BIN_PATH)" | grep -q "$(BUILD_DIR)"; then \
		rm -f "$(BIN_PATH)"; \
		if [ -L "$(BACKUP_PATH)" ]; then \
			mv "$(BACKUP_PATH)" "$(BIN_PATH)"; \
			echo "Restored Homebrew link: $(BIN_PATH) -> $$(readlink "$(BIN_PATH)")"; \
		else \
			echo "Removed: $(BIN_PATH) (no Homebrew version to restore)"; \
		fi; \
	elif [ -L "$(BIN_PATH)" ]; then \
		echo "$(BIN_PATH) points elsewhere, not touching it"; \
	elif [ -f "$(BIN_PATH)" ]; then \
		echo "$(BIN_PATH) is a regular file (Homebrew), not touching it"; \
	else \
		echo "$(BIN_PATH) does not exist"; \
		rm -f "$(BACKUP_PATH)" 2>/dev/null; \
	fi

clean:
	rm -f $(BINARY_NAME) coverage.out

test:
	go test -race ./...

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

tidy:
	go mod tidy

snapshot:
	goreleaser release --snapshot --clean

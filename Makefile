BIN_NAME := bonk
BIN_DIR  := bin
GO       ?= go

VERSION  := 2.0.0
LDFLAGS  := -s -w -X main.version=$(VERSION)

.PHONY: all build build-arm64 build-amd64 build-universal test clean install

all: build

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BIN_NAME) ./cmd/bonk

build-arm64:
	@mkdir -p $(BIN_DIR)
	GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BIN_NAME)_arm64 ./cmd/bonk

build-amd64:
	@mkdir -p $(BIN_DIR)
	CC="clang -arch x86_64" CGO_ENABLED=1 GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BIN_NAME)_amd64 ./cmd/bonk

build-universal: build-arm64 build-amd64
	@mkdir -p $(BIN_DIR)
	lipo -create -output $(BIN_DIR)/$(BIN_NAME)_universal $(BIN_DIR)/$(BIN_NAME)_arm64 $(BIN_DIR)/$(BIN_NAME)_amd64
	@echo "Created universal binary: $(BIN_DIR)/$(BIN_NAME)_universal"

test:
	$(GO) test -v ./...

clean:
	rm -rf $(BIN_DIR) $(BIN_NAME)

install: build
	cp $(BIN_DIR)/$(BIN_NAME) /usr/local/bin/$(BIN_NAME)


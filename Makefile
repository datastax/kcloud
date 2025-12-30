


BIN_DIR := bin
BINARY_NAME := kcloud

default: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/kcloud

install:
	go install ./cmd/kcloud

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)

.PHONY: default build install test clean

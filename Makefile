


default: build

build:
	go build cmd

install:
	go install ./cmd/kcloud

test:
	go test ./...

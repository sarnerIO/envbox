.PHONY: build test lint run clean

BINARY := bin/envbox
VERSION := 0.1.0

build:
	go build -ldflags "-X github.com/spf13/cobra.Command.Version=$(VERSION)" -o $(BINARY) ./cmd/envbox

test:
	go test ./...

run:
	go run ./cmd/envbox

clean:
	rm -rf bin/

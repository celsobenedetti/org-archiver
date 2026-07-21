.PHONY: build install test clean all

BINDIR ?= $(HOME)/.local/bin

all: install

build:
	go build -o bin/ ./cmd/...

install:
	go build -o $(BINDIR)/ ./cmd/...

test:
	go test ./...

clean:
	rm -rf bin

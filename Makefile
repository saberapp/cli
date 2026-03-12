.PHONY: build test fmt lint install clean

BINARY   := saber
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test ./...

fmt:
	gofmt -w .

lint:
	golangci-lint run ./...

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

clean:
	rm -rf bin/

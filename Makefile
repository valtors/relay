BINARY  = RELAY
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -s -w"

.PHONY: build dev test lint clean release

build:
	go build $(LDFLAGS) -o $(BINARY) .

dev:
	go run .

test:
	go test ./... -v -timeout 30s

lint:
	go vet ./...
	test -z "$(shell gofmt -l .)"

clean:
	rm -f $(BINARY) dist/*

release:
	mkdir -p dist
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64  .
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64  .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .

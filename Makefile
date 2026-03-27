.PHONY: build test lint clean

build:
	go build -ldflags "-X main.version=dev -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o emma .

test:
	go test ./... -v -count=1

lint:
	go vet ./...

clean:
	rm -f emma

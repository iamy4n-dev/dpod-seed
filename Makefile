.PHONY: build test generate

build:
	go build -o dpod-seed .

test:
	go test ./...

generate:
	go run ./cmd/generate/main.go

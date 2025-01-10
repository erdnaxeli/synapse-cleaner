all: build

build:
	go build ./cmd/synapse-cleaner.go

style:
	go fmt ./...
	go vet ./...

all: build

build:
	go build -o synapse-cleaner ./cmd

style:
	go fmt ./...
	go vet ./...

server:
	@go build -ldflags="-X main.version=$(shell git describe --tags --always --dirty)" -o server.bin cmd/server/server.go
.PHONY: server

client:
	@go build -ldflags="-X main.version=$(shell git describe --tags --always --dirty)" -o client.bin cmd/client/client.go
.PHONY: client

build: clean server client
.PHONY: build

clean:
	@rm -f server.bin client.bin
.PHONY: clean

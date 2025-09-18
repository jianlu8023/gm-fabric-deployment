VERSION:=$(shell git branch --show-current)-$(shell git describe --tags --always --dirty)
BUILDTIME=$(shell date +"%Y-%m-%d %H:%M:%S")

server:
	@go build -tags=jsoniter -trimpath -ldflags="-s -w -X 'github.com/jianlu8023/golang-example/version.Version=$(VERSION)'" -o server.bin cmd/server/server.go
	@echo -e "version : ${VERSION}\ntime : ${BUILDTIME}" > server.latest
	@echo "server done"
.PHONY: server

client:
	@go build -tags=jsoniter -trimpath -ldflags="-s -w -X 'github.com/jianlu8023/golang-example/version.Version=$(VERSION)'" -o client.bin cmd/client/client.go
	@echo -e "version : ${VERSION}\ntime : ${BUILDTIME}" > client.latest
	@echo "client done"
.PHONY: client

build: clean server client
.PHONY: build

clean:
	@rm -f server.bin client.bin
	@echo "clean done"
.PHONY: clean

IMAGE_VERSION:=v$(shell date +"%Y%m%d%H%M")
DOCKER_FILE:= Dockerfile
IMAGE_NAME:=golang-example/ubuntu2204/app:$(IMAGE_VERSION)

docker:
	@docker pull golang:1.22
	@docker pull ubuntu:22.04
	@docker buildx build --platform linux/amd64 --build-arg "VERSION=${VERSION}" -t "$(IMAGE_NAME)" .
	@docker rmi golang:1.22 ubuntu:22.04
	@docker builder prune -a -f
	@echo "IMAGE NAME: $(IMAGE_NAME)"
	@echo "docker done"
.PHONY: docker

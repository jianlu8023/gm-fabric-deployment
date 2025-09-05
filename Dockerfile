FROM golang:1.22 AS mod-cache

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOSUMDB=sum.golang.google.cn
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-g -O2 -Wno-return-local-addr"

WORKDIR /buildspace

COPY go.mod /buildspace/go.mod
COPY go.sum /buildspace/go.sum

RUN go mod download -x

FROM mod-cache AS go-build

WORKDIR /buildspace

COPY . .

RUN go build -ldflags="-X main.version=$(git describe --tags --always --dirty)" -o server.bin cmd/server/server.go && go build -ldflags="-X main.version=$(shell git describe --tags --always --dirty)" -o client.bin cmd/client/client.go

# 下载grpcurl tini
FROM ubuntu:20.04 AS toolsbuilder

ENV GRPCURL_VERSION=1.9.3
ENV GHPROXY=https://gh-proxy.com/

RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    sed -i s@/security.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y wget curl tini && \
    wget ${GHPROXY}https://github.com/fullstorydev/grpcurl/releases/download/v${GRPCURL_VERSION}/grpcurl_${GRPCURL_VERSION}_linux_amd64.deb && \
    dpkg -i grpcurl_${GRPCURL_VERSION}_linux_amd64.deb


FROM alpine:3.21 AS runner

WORKDIR /myapp

COPY --from=toolsbuilder /usr/bin/tini /usr/bin/tini
COPY --from=toolsbuilder /usr/bin/grpcurl /usr/bin/grpcurl
COPY --from=go-build /buildspace/server.bin /myapp/server.bin
COPY --from=go-build /buildspace/client.bin /myapp/client.bin
COPY --from=go-build /buildspace/configs /myapp/configs
COPY --from=go-build /buildspace/certs /myapp/certs

EXPOSE 8080/tcp \
    65534/tcp \
    65533/tcp

ENTRYPOINT ["/usr/bin/tini","--","/myapp/server.bin"]

CMD ["-config","/myapp/configs/server.yaml","-type","dev"]

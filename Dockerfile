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

ARG VERSION

ENV VERSION=${VERSION}




COPY . .

RUN echo "starting build server.bin" && \
    go build -tags='jsoniter' -trimpath -ldflags="-s -w -X 'github.com/jianlu8023/golang-example/version.Version=${VERSION}'" -o server.bin cmd/server/server.go && \
    echo "starting build client.bin" && \
    go build -tags='jsoniter' -trimpath -ldflags="-s -w -X 'github.com/jianlu8023/golang-example/version.Version=${VERSION}'" -o client.bin cmd/client/client.go

# 下载grpcurl tini
FROM ubuntu:22.04 AS toolsbuilder

ENV GRPCURL_VERSION=1.9.3
ENV GHPROXY=https://ghproxy.8023202.xyz/

RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    sed -i s@/security.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y wget curl tini && \
    wget ${GHPROXY}https://github.com/fullstorydev/grpcurl/releases/download/v${GRPCURL_VERSION}/grpcurl_${GRPCURL_VERSION}_linux_amd64.deb && \
    dpkg -i grpcurl_${GRPCURL_VERSION}_linux_amd64.deb


# FROM alpine:3.21 AS runner
FROM ubuntu:22.04 AS runner

#ENV GOROOT=/usr/local/go
#RUN mkdir -p $GOROOT/lib/time
#COPY --from=go-build /usr/local/go/lib/time/zoneinfo.zip $GOROOT/lib/time/zoneinfo.zip

ENV DEBIAN_FRONTEND=noninteractive

RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    sed -i s@/security.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    apt-get update && \
    apt-get install curl tzdata gosu tini -y && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    dpkg-reconfigure -f noninteractive tzdata && \
    apt-get clean cache && \
    apt-get autoremove -y && \
    apt-get autoclean && \
    rm -rf /var/lib/apt/lists/*


WORKDIR /myapp

COPY --from=toolsbuilder /usr/bin/tini /usr/bin/tini
COPY --from=toolsbuilder /usr/bin/grpcurl /usr/bin/grpcurl
COPY --from=go-build /buildspace/server.bin /myapp/server.bin
COPY --from=go-build /buildspace/client.bin /myapp/client.bin
COPY --from=go-build /buildspace/configs /myapp/configs
COPY --from=go-build /buildspace/certs /myapp/certs
COPY --from=go-build /buildspace/docker-entrypoint.sh /docker-entrypoint.sh

#创建非 root 用户
RUN mkdir -p /myapp/logs && \
    mkdir -p /myapp/db && \
    groupadd -r myusers && useradd -r -u 1000 -g myusers appuser && \
    chown -R appuser:myusers /myapp && \
    chown appuser:myusers /docker-entrypoint.sh && \
    chmod +x /docker-entrypoint.sh

# 需要docker.sock
USER root

EXPOSE 8080/tcp \
    65534/tcp \
    65533/tcp \
    2000/tcp \
    2001/tcp

VOLUME /myapp/logs \
    /myapp/db \
    /myapp/certs \
    /myapp/configs


ENTRYPOINT ["/usr/bin/tini","--","/docker-entrypoint.sh"]

HEALTHCHECK --interval=60s --timeout=5s --retries=3 --start-period=30s CMD curl -ksS https://localhost:8080/example/health || exit 1

CMD ["server","--config","/myapp/configs/server.yaml","--type","dev"]

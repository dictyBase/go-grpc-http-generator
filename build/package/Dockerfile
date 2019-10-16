## build prototool and go tools
FROM dictybase/golang:1.12.10-alpine-edge AS protobuilder
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
    && apk add --no-cache build-base curl git protobuf \
    && rm -rf /var/cache/apk/*

ENV GOLANG_PROTOBUF_VERSION=1.3.2
ENV PROTOTOOL_VERSION=1.8.0
ENV GO_PROTO_VALIDATORS=0.2.0
ENV PROTOC_GEN_DOC=1.3.0
ENV HUB_VERSION=2.12.8
ENV GO111MODULE=on
# build all protocol buffer prerequisites
RUN go get -v github.com/golang/protobuf/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} \
    && go get -v github.com/uber/prototool/cmd/prototool@v${PROTOTOOL_VERSION} \
    && go get -v github.com/mwitkow/go-proto-validators/protoc-gen-govalidators@v${GO_PROTO_VALIDATORS} \
    && go get -v github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v${PROTOC_GEN_DOC} \
    && go get -v github.com/github/hub@v${HUB_VERSION}

# build the codebase
RUN mkdir -p /go-grpc-http-generator
WORKDIR /go-grpc-http-generator
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o app

# Runner image
FROM dictybase/golang:1.11.13-alpine-edge
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
    && apk add --no-cache ca-certificates protobuf tzdata curl \
    && rm -rf /var/cache/apk/* \
    && cp /usr/share/zoneinfo/America/Chicago /etc/localtime \
    && echo "America/chicago" > /etc/timezone

ENV GOPROXY https://proxy.golang.org
## install grpc-gateway binaries
ENV GRPC_GATEWAY_VERSION=1.11.3
RUN curl -sSL \
    https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-grpc-gateway-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
    -o /usr/local/bin/protoc-gen-grpc-gateway \
    && curl -sSL \
    https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-swagger-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
    -o /usr/local/bin/protoc-gen-swagger \
    && chmod +x /usr/local/bin/protoc-gen-grpc-gateway \
    && chmod +x /usr/local/bin/protoc-gen-swagger

## configure protocol buffer tools
COPY --from=protobuilder /go/bin/protoc-gen* /usr/local/bin/
COPY --from=protobuilder /go/bin/prototool /usr/local/bin/
## copy hub binary
COPY --from=protobuilder /go/bin/hub /usr/local/bin/
## copy app
COPY --from=protobuilder /go-grpc-http-generator/app /usr/local/bin/

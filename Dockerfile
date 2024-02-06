FROM golang:1.21.7-alpine AS build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GOOS=linux
ENV GOARCH=amd64

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}

RUN apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /go/src/github.com/larivierec/cloudflare-ddns
COPY . /go/src/github.com/larivierec/cloudflare-ddns/

RUN go mod download
RUN go build -o ddns cmd/ddns.go

FROM gcr.io/distroless/static:nonroot
USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /go/src/github.com/larivierec/cloudflare-ddns/ddns /usr/local/bin/ddns
COPY --from=build --chown=nonroot:nonroot /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT [ "ddns" ]

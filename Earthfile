VERSION 0.6
FROM golang:1.19-bullseye
WORKDIR /jelease

deps:
  COPY go.mod go.sum ./
  RUN go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
  FROM +deps
  COPY *.go .
  RUN go test -v ./... \
    && go build -o build/jelease main.go
  SAVE ARTIFACT build/jelease /jelease AS LOCAL build/jelease

docker:
  FROM ubuntu:22.04
  RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
  COPY +build/jelease .
  CMD ["/jelease"]
  SAVE IMAGE jelease:latest
  SAVE IMAGE --push docker-riskident.2rioffice.com/platform/jelease

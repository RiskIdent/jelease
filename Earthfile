# SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
#
# SPDX-License-Identifier: CC0-1.0

VERSION 0.6
FROM golang:1.21-bullseye
WORKDIR /jelease

deps:
  COPY go.mod go.sum ./
  RUN go mod download
  SAVE ARTIFACT go.mod AS LOCAL go.mod
  SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
  ARG VERSION=devel
  FROM +deps
  COPY . .
  RUN go test -v ./... \
    && go build \
      -ldflags "-X github.com/RiskIdent/jelease/cmd.appVersion=$VERSION" \
      -o build/jelease main.go
  SAVE ARTIFACT build/jelease /jelease AS LOCAL build/jelease

docker:
  ARG VERSION=latest
  ARG REGISTRY=ghcr.io/riskident
  FROM ubuntu:22.04
  RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git \
    && rm -rf /var/lib/apt/lists/*
  COPY +build/jelease /usr/local/bin
  CMD ["jelease", "serve"]
  LABEL org.opencontainers.image.source=https://github.com/RiskIdent/jelease
  LABEL org.opencontainers.image.description="Automatically create software update PRs and Jira tickets using newreleases.io webhooks."
  LABEL org.opencontainers.image.licenses=GPL-3.0-or-later
  IF [ "$VERSION" != "latest" ]
    SAVE IMAGE --push $REGISTRY/jelease:latest
  END
  SAVE IMAGE --push $REGISTRY/jelease:$VERSION

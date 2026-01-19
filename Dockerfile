# SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
#
# SPDX-License-Identifier: CC0-1.0

FROM docker.io/library/golang:1.26rc2-alpine AS build
WORKDIR /jelease
COPY go.mod go.sum ./
RUN go mod download

ARG VERSION=latest
COPY . .
RUN go test -v ./... \
  && go build \
    -ldflags "-X github.com/RiskIdent/jelease/cmd.appVersion=$VERSION" \
    -o build/jelease main.go

# NOTE: When updating here, remember to also update in ./goreleaser.Dockerfile
FROM docker.io/library/alpine AS final
RUN apk add --no-cache ca-certificates diffutils patch git git-lfs helm \
  && addgroup -g 10000 jelease \
  && adduser -D -u 10000 -G jelease jelease
COPY --from=build /jelease/build/jelease /usr/local/bin/
CMD ["jelease", "serve"]
USER 10000
LABEL org.opencontainers.image.source=https://github.com/RiskIdent/jelease
LABEL org.opencontainers.image.description="Automatically create software update PRs and Jira tickets using newreleases.io webhooks."
LABEL org.opencontainers.image.licenses=GPL-3.0-or-later

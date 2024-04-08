# SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
#
# SPDX-License-Identifier: CC0-1.0

FROM alpine
RUN apk add --no-cache ca-certificates patch git git-lfs helm
COPY jelease /usr/local/bin/
CMD ["jelease", "serve"]
USER 10000
LABEL org.opencontainers.image.source=https://github.com/RiskIdent/jelease
LABEL org.opencontainers.image.description="Automatically create software update PRs and Jira tickets using newreleases.io webhooks."
LABEL org.opencontainers.image.licenses=GPL-3.0-or-later

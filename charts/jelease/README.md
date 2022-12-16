<!--
SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>

SPDX-License-Identifier: CC-BY-4.0
-->

# Jelease helm chart

Deploys [github.com/RiskIdent/jelease](https://github.com/RiskIdent/jelease)
with a [webhookrelay](https://webhookrelay.com/) sidecar to Kubernetes.

## Deploy

```bash
helm upgrade --install jelease . --namespace jelease --create-namespace
```

Encrypt the secret values before committing them using
[git-crypt](https://github.com/AGWA/git-crypt)
or [SOPS](https://github.com/mozilla/sops),
or store them in a different secure location.

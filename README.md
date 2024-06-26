<!--
SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>

SPDX-License-Identifier: CC-BY-4.0
-->

<p align="center">
  <img src="./docs/jelease-gopher-card-512.jpg" alt="jelease gopher logo"/>
</p>

<h1 align="center">jelease - A newreleases.io ➡️ Jira connector</h1>

[![REUSE status](https://api.reuse.software/badge/github.com/RiskIdent/jelease)](https://api.reuse.software/info/github.com/RiskIdent/jelease)

Automatically create GitHub PRs on package updates but with more control.

## How it works

Jelease first creates a Jira issue telling you that you need to
update your software.

Then if Jelease has configuration on how to update the package for you,
it will go ahead and create PRs.

```mermaid
flowchart LR
    newreleases.io -->|update event| Jelease
    Jelease --> jira[/Create Jira issue/]
    Jelease --> config{know how to\ncreate PR}
    config -->|yes| pr[/Create GitHub PR/]
```

When hosting Jelease inside an internal network, you can make use of
services like [Webhook Relay](https://webhookrelay.com/), so you
don't need to expose Jelease to the internet.

```mermaid
flowchart LR
    subgraph Internet
      newreleases.io
      webhookrelay[Webhook Relay]
    end
    newreleases.io -->|update event| webhookrelay
    webhookrelayd --->|pull events| webhookrelay
    subgraph internal network
      Jelease --> jira[/Create Jira issue/]
      Jelease --> pr[/Create GitHub PR/]
      webhookrelayd -->|push event| Jelease
    end
```

## Motivation

We wanted our software to stay up to date, and in our
Infrastructure as Code (IaC) practices we couldn't use conventional tools like
[Dependabot](https://docs.github.com/en/code-security/dependabot)
or [Renovate](https://www.mend.io/renovate/). Sample use cases:

- Ansible host_vars specifying package versions
- Helm chart values specifying Docker image tags

At our workplace, we still use Dependabot, but then fill the gaps with Jelease.

## Configuration

Jelease is configured via YAML files.
See [`jelease.yaml`](./jelease.yaml) for the default values.

### Configuration files

Jelease reads from the first config file in this list that it can find:

| Order | Linux                       | Mac                                          | Windows                       |
| ----: | --------------------------- | -------------------------------------------- | ----------------------------- |
|    1. | `/etc/jelease/jelease.yaml` | `/etc/jelease/jelease.yaml`                  | *N/A*                         |
|    2. | `~/.config/jelease.yaml`    | `~/Library/Application Support/jelease.yaml` | `%APPDATA%\jelease.yaml`      |
|    3. | `~/.jelease.yaml`           | `~/.jelease.yaml`                            | `%USERPROFILE%\.jelease.yaml` |
|    4. | `./jelease.yaml`            | `./jelease.yaml`                             | `.\jelease.yaml`              |

### Package config

In the configuration you can specify "packages", which tells Jelease how to
create PRs. The package configs are just a list of repositories to touch,
followed by which "patches" it should perform. The available patches are:

- `regex`: Do a search and replace inside a file.

- `yaml`: Use YAML Path (similar to JSON Path)
  to target a specific field to update

- `helmDepUpdate`: Run `helm dep update` inside a directory.

In these configs we allow you to template a lot of values using Go templates.
All templates allow you to use the following values:

- `{{ .Package }}` resolves to the name of the package, e.g `RiskIdent/jelease`

- `{{ .Version }}` resolves to the version of the package, e.g `v0.6.2`

- `{{ .JiraIssue }}` resolves to the Jira issue key, e.g `TICKET-1234`,
  or empty if no issue was created (such as during dry runs)

### JSON Schema

There's also a [JSON Schema](https://json-schema.org/) for the config file,
which gives you warnings and completion support inside your IDE.

Make use of it via e.g:

- [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)
  for [VS Code](https://code.visualstudio.com/).

- [coc-yaml plugin](https://github.com/neoclide/coc-yaml)
  for [coc.nvim](https://github.com/neoclide/coc.nvim),
  an extension framework for both Vim and NeoVim.

To make use of it, add the following comment to the beginning of your
config file:

```yaml
# yaml-language-server: $schema=https://github.com/RiskIdent/jelease/raw/main/jelease.schema.json
```

## Local usage

1. Create a GitHub PAT (e.g on <https://github.com/settings/tokens>)

2. Create a local config file, e.g at `~/.config/jelease.yaml`

3. Add your local config to test with, including your newly generated PAT. E.g:

   ```yaml
   # yaml-language-server: $schema=https://github.com/RiskIdent/jelease/raw/main/jelease.schema.json

   packages:
     - name: neuvector
       repos:
         - url: https://github.example.com/some-org/some-repo
           patches:
             - regex:
                 file: helm/charts/ri-neuvector/Chart.yaml
                 match: '^appVersion: .*'
                 replace: 'appVersion: {{ .Version }}'
             - regex:
                 file: helm/charts/ri-neuvector/Chart.yaml
                 match: '^version: (.*)'
                 replace: 'version: {{ index .Groups 1 | versionBump "0.0.1" }}'

   github:
     url: https://github.example.com

     auth:
       type: pat
       token: ghp_loremipsum
   ```

4. Test that your local config is read correctly:

   ```bash
   go run . config
   ```

5. Run Jelease locally, e.g:

   ```bash
   # To test applying changes
   go run . apply neuvector v1.2.3 --dryrun

   # To test creating PRs
   go run . apply neuvector v1.2.3

   # To test the webhook receiver server, without creating issues or PRs
   go run . serve --dryrun

   # To test the webhook receiver server, with creating issues and PRs
   go run . serve
   ```

6. To test out the webhooks, you can make use of our example webhook like so:

   ```bash
   curl localhost:8080/webhook -d @examples/newreleasesio-webhook.json
   ```

## Development

Prerequisites:

- Go 1.21 (or later): <https://go.dev/doc/install>

The HTML pages are templated using Templ. This is a manual step that
needs to be done after each code change to any `.templ` files.

```bash
go generate
```

The generate command effectively runs `go run ...`, so you don't
need `templ` CLI installed.

## Building the application and docker image

```bash
VERSION=v0.4.1

podman build . -t ghcr.io/riskident/jelease:${VERSION} --build-arg VERSION=${VERSION}
```

## Releasing

1. Create a new release on GitHub, with "v" prefix on version: <https://github.com/RiskIdent/jelease/releases/new>

2. Write a small changelog, like so:

   ```markdown
   ## Changes (since v0.3.0)

   - Added some feature. (#123)
   ```

3. Our GitHub Action with goreleaser will build and add artifacts to release

## Logo

The gopher logo is designed by Kristin Weyand, an employee at [Risk.Ident](https://riskident.com).

The gopher logo of Jelease was inspired by the original Go gopher,
designed by [Renee French](https://reneefrench.blogspot.com/).

## License

This repository complies with the [REUSE recommendations](https://reuse.software/).

Different licenses are used for different files. In general:

- Go code is licensed under GNU General Public License v3.0 or later ([LICENSES/GPL-3.0-or-later.txt](LICENSES/GPL-3.0-or-later.txt)).
- Documentation licensed under Creative Commons Attribution 4.0 International ([LICENSES/CC-BY-4.0.txt](LICENSES/CC-BY-4.0.txt)).
- The logo is licensed under Creative Commons Attribution 4.0 International ([LICENSES/CC-BY-4.0.txt](LICENSES/CC-BY-4.0.txt)).
- Miscellaneous files, e.g `.gitignore`, are licensed under CC0 1.0 Universal ([LICENSES/CC0-1.0.txt](LICENSES/CC0-1.0.txt)).

Please see each file's header or accompanied `.license` file for specifics.

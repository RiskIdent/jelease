# jelease - A newreleases.io ➡️ Jira connector

![logo](./docs/new-release-128.png)

Automatically create Jira tickets when a newreleases.io release
is detected using webhooks.

## Configuration

The application requires the following environment variables to be set:
- Connection and authentication

  - `JELEASE_AUTH_TYPE`: One of [pat, token]. Determines whether to authenticate using personal access token (on premise) or jira api token (jira cloud)
  - `JELEASE_JIRA_TOKEN`: Jira API token, can also be a password in self-hosted instances
  - `JELEASE_JIRA_URL`: The URL of your Jira instance
  - `JELEASE_JIRA_USER`: Jira username to authenticate API requests
  - `JELEASE_PORT`: The port the application is expecting traffic on
  - `JELEASE_INSECURE_SKIP_CERT_VERIFY`: Skips verification of Jira server certs when performing http requests.
- Jira ticket creation:
  - `JELEASE_ADD_LABELS`: Comma-separated list of labels to add to the created jira ticket
  - `JELEASE_DEFAULT_STATUS`: The status the created tickets are supposed to have
  - `JELEASE_DRY_RUN`: Don't create tickets, log when a ticket would be created
  - `JELEASE_ISSUE_DESCRIPTION`: The description for created issues
  - `JELEASE_PROJECT`: Jira Project key the tickets will be created in
  - `JELEASE_LOG_FORMAT`: Logging format. One of: `pretty` (default), `json`
  - `JELEASE_LOG_LEVEL`: Logging minimum level/severity. One of: `trace`, `debug` (default), `info`, `warn`, `error`, `fatal`, `panic`

They can also be specified using a `.env` file in the application directory.

## Local usage

1. Populate a `.env` file with configuration values
2. `go run main.go` / `./jelease`
3. Direct newreleases.io webhooks to the `host:port/webhook` route.

## Building the application and docker image

The application uses [earthly](https://earthly.dev/get-earthly) for building
and pushing a docker image.

After installing earthly, the image can be built by running

```bash
earthly +docker
# if you want to push a new image version
earhtly --push +docker
```

## Deployment

A helm chart deploying the application together with a webhookrelayd sidecar
is available in the
[platform/helm repo](https://github.2rioffice.com/platform/helm/tree/master/charts/jelease)

## Logo

[New-release icons created by berkahicon - Flaticon](https://www.flaticon.com/free-icons/new-release)

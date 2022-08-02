# jelease - A newreleases.io ➡️ Jira connector

Automatically create Jira tickets when a newreleases.io release is detected using webhooks.

## Configuration:

The application requires the following environment variables to be set:
- `JELEASE_PORT`: The port the application is expecting traffic on
- `JELEASE_JIRA_URL`: The URL of your Jira instance
- `JELEASE_JIRA_USER`: Jira username to authenticate API requests
- `JELEASE_JIRA_TOKEN`: Jira API token, can also be a password in self-hosted instances
- `JELEASE_PROJECT`: Jira Project key the tickets will be created in
- `JELEASE_ADD_LABELS`: Add additional labels to the created jira ticket
- `JELEASE_DEFAULT_STATUS`: The status the created tickets are supposed to have

They can also be specified using a `.env` file in the application directory.

## Building the application

`go build`

## Usage

1. `go run main.go` / `./jelease`
2. Direct newreleases.io webhooks to the `host:port/webhook` route.
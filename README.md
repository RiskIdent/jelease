# jelease - A newreleases.io ➡️ Jira connector

Automatically create Jira tickets when a newreleases.io release is detected using webhooks.

## Configuration:

The application requires the following environment variables to be set:
- `PORT`: The port the application is expecting traffic on
- `JIRA_URL`: The URL of your Jira instance
- `JIRA_USER`: Jira username to authenticate API requests
- `JIRA_TOKEN`: Jira API token, can also be a password in self-hosted instances
- `JIRA_PROJECT`: Jira Project key the tickets will be created in

## Usage

Send newreleases.io webhooks to the `host:port/webhook` route.
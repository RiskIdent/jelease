# jelease - A newreleases.io ➡️ Jira connector

Automatically create Jira tickets when a newreleases.io release is detected using webhooks.

## Configuration:

The application requires the following environment variables to be set:
- `JELEASE_PORT`: The port the application is expecting traffic on
- `JELEASE_ADDLABELS`: Add additional labels to the created jira ticket
- `JELEASE_JIRAURL`: The URL of your Jira instance
- `JELEASE_JIRAUSER`: Jira username to authenticate API requests
- `JELEASE_JIRATOKEN`: Jira API token, can also be a password in self-hosted instances
- `JELEASE_JIRAPROJECT`: Jira Project key the tickets will be created in

## Usage

Direct newreleases.io webhooks to the `host:port/webhook` route.
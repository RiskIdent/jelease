package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/trivago/tgo/tcontainer"
)

var (
	jiraClient *jira.Client
	config     Config
)

// Config contains configuration values from environment and .env file.
// Environment takes precedence over the .env file in case of conflicts.
type Config struct {
	// connection and auth
	AuthType       string `envconfig:"JELEASE_AUTH_TYPE" required:"true"`
	JiraToken      string `envconfig:"JELEASE_JIRA_TOKEN" required:"true"`
	JiraURL        string `envconfig:"JELEASE_JIRA_URL" required:"true"`
	JiraUser       string `envconfig:"JELEASE_JIRA_USER" required:"true"`
	Port           int    `envconfig:"JELEASE_PORT" default:"8080"`
	SkipCertVerify bool   `envconfig:"JELEASE_INSECURE_SKIP_CERT_VERIFY" default:"false"`

	// ticket creation
	AddLabels              []string `envconfig:"JELEASE_ADD_LABELS"`
	DefaultStatus          string   `envconfig:"JELEASE_DEFAULT_STATUS" required:"true"`
	DryRun                 bool     `envconfig:"JELEASE_DRY_RUN" default:"false"`
	IssueDescription       string   `envconfig:"JELEASE_ISSUE_DESCRIPTION" default:"Update issue generated by https://github.2rioffice.com/platform/jelease using newreleases.io"`
	IssueType              string   `envconfig:"JELEASE_ISSUE_TYPE" default:"Story"`
	Project                string   `envconfig:"JELEASE_PROJECT" required:"true"`
	ProjectNameCustomField uint     `envconfig:"JELEASE_PROJECT_NAME_CUSTOM_FIELD"`

	// logging
	LogFormat string `envconfig:"JELEASE_LOG_FORMAT" default:"pretty"`
	LogLevel  string `envconfig:"JELEASE_LOG_LEVEL" default:"debug"`
}

// Release object unmarshaled from the newreleases.io webhook.
// Some fields omitted for simplicity, refer to the documentation at https://newreleases.io/webhooks
type Release struct {
	Provider string `json:"provider"`
	Project  string `json:"project"`
	Version  string `json:"version"`
}

// Generates a Textual summary for the release, intended to be used as the Jira issue summary
func (r Release) IssueSummary() string {
	return fmt.Sprintf("Update %v to version %v", r.Project, r.Version)
}

func (r Release) JiraIssue() jira.Issue {
	labels := config.AddLabels
	var extraFields tcontainer.MarshalMap

	if config.ProjectNameCustomField == 0 {
		labels = append(labels, r.Project)
	} else {
		customFieldName := fmt.Sprintf("customfield_%d", config.ProjectNameCustomField)
		extraFields = tcontainer.MarshalMap{
			customFieldName: r.Project,
		}
	}
	return jira.Issue{
		Fields: &jira.IssueFields{
			Description: config.IssueDescription,
			Project: jira.Project{
				Key: config.Project,
			},
			Type: jira.IssueType{
				Name: config.IssueType,
			},
			Labels:   labels,
			Summary:  r.IssueSummary(),
			Unknowns: extraFields,
		},
	}
}

// handleGetRoot handles to GET requests for a basic reachability check
func handleGetRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Ok")
}

// handlePostWebhook handles newreleases.io webhook post requests
func handlePostWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		status := http.StatusMethodNotAllowed
		statusText := http.StatusText(status)
		http.Error(w, statusText, status)
		log.Debug().
			Str("method", r.Method).
			Int("status", status).
			Str("statusText", statusText).
			Msg("Rejected request, only POST allowed.")
		return
	}

	if log.Logger.GetLevel() == zerolog.TraceLevel {
		start := time.Now()
		defer func() {
			log.Trace().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Dur("dur", time.Since(start)).
				Msg("Received request.")
		}()
	}

	// parse newreleases.io webhook
	decoder := json.NewDecoder(r.Body)
	var release Release
	err := decoder.Decode(&release)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		log.Error().Err(err).Msg("Failed to decode request body to JSON.")
		return
	}

	// look for existing update tickets
	existingIssuesQuery := newJiraIssueSearchQuery(config.DefaultStatus, release.Project, config.ProjectNameCustomField)
	existingIssues, resp, err := jiraClient.Issue.Search(existingIssuesQuery, &jira.SearchOptions{})
	if err != nil {
		err := fmt.Errorf("searching Jira for previous issues: %w", err)
		logJiraErrResponse(resp, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(existingIssues) == 0 {
		// no previous issues, create new jira issue
		i := release.JiraIssue()
		log.Trace().Interface("issue", i).Msg("Creating issue.")
		if config.DryRun {
			log.Debug().
				Str("issue", i.Fields.Summary).
				Msg("Skipping creation of issue because Config.DryRun is enabled.")
			return
		}
		newIssue, resp, err := jiraClient.Issue.Create(&i)
		if err != nil {
			err := fmt.Errorf("creating Jira issue: %w", err)
			logJiraErrResponse(resp, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Info().Str("issue", newIssue.Key).Msg("Created issue.")
		return
	}

	// in case of duplicate issues, update the oldest (probably original) one, ignore rest as duplicates
	var oldestExistingIssue jira.Issue
	var duplicateIssueKeys []string
	for i, existingIssue := range existingIssues {
		if i == 0 {
			oldestExistingIssue = existingIssue
			continue
		}
		tCurrent := time.Time(existingIssue.Fields.Created)
		tOldest := time.Time(oldestExistingIssue.Fields.Created)
		if tCurrent.Before(tOldest) {
			duplicateIssueKeys = append(duplicateIssueKeys, oldestExistingIssue.Key)
			oldestExistingIssue = existingIssue
		} else {
			duplicateIssueKeys = append(duplicateIssueKeys, existingIssue.Key)
		}
	}
	if len(duplicateIssueKeys) > 0 {
		log.Debug().
			Str("older", oldestExistingIssue.Key).
			Strs("duplicates", duplicateIssueKeys).
			Msg("Ignoring the following possible duplicate issues in favor of older issue.")
	}

	// This seems hacky, but is taken from the official examples
	// https://github.com/andygrunwald/go-jira/blob/47d27a76e84da43f6e27e1cd0f930e6763dc79d7/examples/addlabel/main.go
	// There is also a jiraClient.Issue.Update() method, but it panics and does not provide a usage example
	type summaryUpdate struct {
		Set string `json:"set" structs:"set"`
	}
	type issueUpdate struct {
		Summary []summaryUpdate `json:"summary" structs:"summary"`
	}
	previousSummary := oldestExistingIssue.Fields.Summary
	if config.DryRun {
		log.Debug().
			Str("issue", oldestExistingIssue.Key).
			Str("summary", release.IssueSummary()).
			Msg("Skipping update of issue because Config.DryRun is enabled.")
		return
	}
	updates := map[string]any{
		"update": issueUpdate{
			Summary: []summaryUpdate{
				{Set: release.IssueSummary()},
			},
		},
	}
	log.Trace().Interface("updates", updates).Msg("Updating issue.")
	resp, err = jiraClient.Issue.UpdateIssue(oldestExistingIssue.ID, updates)
	if err != nil {
		err := fmt.Errorf("update Jira issue: %w", err)
		logJiraErrResponse(resp, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Info().
		Str("issue", oldestExistingIssue.Key).
		Str("from", previousSummary).
		Str("to", release.IssueSummary()).
		Msg("Updated issue summary")
}

func configSetup() error {
	dotEnvErr := godotenv.Load()

	err := envconfig.Process("jelease", &config)
	if err != nil {
		loggerSetup() // need to be setup before returning
		return err
	}
	if err := loggerSetup(); err != nil {
		return err
	}

	if os.IsNotExist(dotEnvErr) {
		log.Debug().Msg("No .env file found.")
	} else if dotEnvErr != nil {
		return dotEnvErr
	}

	log.Debug().Str("url", config.JiraURL).Msg("Loaded configuration.")

	var httpClient *http.Client
	tlsConfig := tls.Config{InsecureSkipVerify: config.SkipCertVerify}

	switch strings.ToLower(config.AuthType) {
	case "pat":
		httpClient = (&jira.PATAuthTransport{
			Token:     config.JiraToken,
			Transport: &http.Transport{TLSClientConfig: &tlsConfig},
		}).Client()
	case "token":
		httpClient = (&jira.BasicAuthTransport{
			Username:  config.JiraUser,
			Password:  config.JiraToken,
			Transport: &http.Transport{TLSClientConfig: &tlsConfig},
		}).Client()
	default:
		return fmt.Errorf("invalid AUTH_TYPE value %q, must be one of [pat, token]", config.AuthType)
	}

	httpClient.Timeout = 10 * time.Second
	jiraClient, err = jira.NewClient(httpClient, config.JiraURL)
	if err != nil {
		return fmt.Errorf("create jira client: %w", err)
	}

	return nil
}

func projectExists() error {
	allProjects, response, err := jiraClient.Project.GetList()
	if err != nil {
		errCtx := errors.New("error response from Jira when retrieving project list")
		if response != nil {
			body, readErr := io.ReadAll(response.Body)
			if readErr != nil {
				return fmt.Errorf("%v: %w. Failed to decode response body: %v", errCtx, err, readErr)
			}
			return fmt.Errorf("%v: %w. Response body: %v", errCtx, err, string(body))
		}
		return fmt.Errorf("%v: %w", errCtx, err)
	}
	var projectExists bool
	for _, project := range *allProjects {
		if project.Key == config.Project {
			projectExists = true
			break
		}
	}
	if !projectExists {
		return fmt.Errorf("project %v does not exist on your Jira server", config.Project)
	}
	return nil
}

func statusExists() error {
	allStatuses, response, err := jiraClient.Status.GetAllStatuses()
	if err != nil {
		errCtx := errors.New("error response from Jira when retrieving status list: %+v")
		if response != nil {
			body, readErr := io.ReadAll(response.Body)
			if readErr != nil {
				return fmt.Errorf("%v: %w. Failed to decode response body: %v", errCtx, err, readErr)
			}
			return fmt.Errorf("%v: %w. Response body: %v", errCtx, err, string(body))
		}
		return fmt.Errorf("%v: %w", errCtx, err)
	}
	var statusExists bool
	for _, status := range allStatuses {
		if status.Name == config.DefaultStatus {
			statusExists = true
			break
		}
	}
	if !statusExists {
		var statusSB strings.Builder
		for i, status := range allStatuses {
			if i > 0 {
				statusSB.WriteString(", ")
			}
			statusSB.WriteString(status.Name)
		}
		return fmt.Errorf("status %q does not exist on your Jira server for project %q. Available statuses: [%v]", config.DefaultStatus, config.Project, statusSB.String())
	}
	return nil
}

func serveHTTP() error {
	http.HandleFunc("/webhook", handlePostWebhook)
	http.HandleFunc("/", handleGetRoot)
	log.Info().Int("port", config.Port).Msg("Starting server.")
	return http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil)
}

func main() {
	err := run()
	if errors.Is(err, http.ErrServerClosed) {
		log.Error().Msg("Server closed.")
	} else if err != nil {
		log.Error().Err(err).Msg("Error starting server.")
		os.Exit(1)
	}
}

func run() error {
	err := configSetup()
	if err != nil {
		return fmt.Errorf("config setup: %w", err)
	}
	err = projectExists()
	if err != nil {
		return fmt.Errorf("check if configured project exists: %w", err)
	}
	log.Debug().Str("project", config.Project).Msg("Configured project found ✓")
	err = statusExists()
	if err != nil {
		return fmt.Errorf("check if configured default status exists: %w", err)
	}
	log.Debug().Str("status", config.DefaultStatus).Msg("Configured default status found ✓")
	return serveHTTP()
}

func loggerSetup() error {
	pretty := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "Jan-02 15:04",
	})
	switch strings.ToLower(config.LogFormat) {
	case "json":
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	case "pretty":
		log.Logger = pretty
	default:
		log.Logger = pretty
		return fmt.Errorf("invalid log format %q, must be one of [json, pretty]", config.LogFormat)
	}

	if config.LogLevel != "" {
		level, err := zerolog.ParseLevel(config.LogLevel)
		if err != nil {
			return err
		}
		log.Logger = log.Level(level)
	}
	return nil
}

func newJiraIssueSearchQuery(statusName, projectName string, customFieldID uint) string {
	if customFieldID == 0 {
		return fmt.Sprintf("status = %q and labels = %q", statusName, projectName)
	}
	// Checking label as well for backward compatibility
	return fmt.Sprintf("status = %q and (labels = %q or cf[%d] ~ %[2]q)",
		statusName, projectName, customFieldID)
}

func logJiraErrResponse(resp *jira.Response, err error) {
	if resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Error().Err(err).Msg("Failed Jira request, and to decode Jira response body.")
		} else {
			var obj any
			if err := json.Unmarshal(body, &obj); err != nil {
				log.Error().Err(err).Str("body", string(body)).Msg("Failed Jira request.")
			} else {
				log.Error().Err(err).Interface("body", obj).Msg("Failed Jira request.")
			}
		}
	} else {
		log.Error().Err(err).Msg("Failed to create Jira issue.")
	}
}

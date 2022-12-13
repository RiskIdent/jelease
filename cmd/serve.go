// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

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

	"github.com/RiskIdent/jelease/pkg/config"
	jira "github.com/andygrunwald/go-jira"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/trivago/tgo/tcontainer"
)

var serveCmd = &cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		err := run()
		if errors.Is(err, http.ErrServerClosed) {
			log.Error().Msg("Server closed.")
		} else if err != nil {
			log.Error().Err(err).Msg("Error starting server.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func run() error {
	jiraClient, err := jiraClientSetup()
	if err != nil {
		return fmt.Errorf("create jira client: %w", err)
	}
	err = projectExists(jiraClient)
	if err != nil {
		return fmt.Errorf("check if configured project exists: %w", err)
	}
	log.Debug().Str("project", cfg.Jira.Issue.Project).Msg("Configured project found ✓")
	err = statusExists(jiraClient)
	if err != nil {
		return fmt.Errorf("check if configured default status exists: %w", err)
	}
	log.Debug().Str("status", cfg.Jira.Issue.Status).Msg("Configured default status found ✓")
	return serveHTTP(jiraClient)
}

func jiraClientSetup() (*jira.Client, error) {
	var httpClient *http.Client
	tlsConfig := tls.Config{InsecureSkipVerify: cfg.Jira.SkipCertVerify}

	switch cfg.Jira.Auth.Type {
	case config.JiraAuthTypePAT:
		httpClient = (&jira.PATAuthTransport{
			Token:     cfg.Jira.Auth.Token,
			Transport: &http.Transport{TLSClientConfig: &tlsConfig},
		}).Client()
	case config.JiraAuthTypeToken:
		httpClient = (&jira.BasicAuthTransport{
			Username:  cfg.Jira.Auth.User,
			Password:  cfg.Jira.Auth.Token,
			Transport: &http.Transport{TLSClientConfig: &tlsConfig},
		}).Client()
	default:
		return nil, fmt.Errorf("invalid Jira auth type %q", cfg.Jira.Auth.Type)
	}

	httpClient.Timeout = 10 * time.Second
	jiraClient, err := jira.NewClient(httpClient, cfg.Jira.URL)
	if err != nil {
		return nil, err
	}
	return jiraClient, nil
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
	labels := cfg.Jira.Issue.Labels
	var extraFields tcontainer.MarshalMap

	if cfg.Jira.Issue.ProjectNameCustomField == 0 {
		log.Trace().Msg("Create ticket with project name in labels.")
		labels = append(labels, r.Project)
	} else {
		log.Trace().
			Uint("customField", cfg.Jira.Issue.ProjectNameCustomField).
			Msg("Create ticket with project name in custom field.")
		customFieldName := fmt.Sprintf("customfield_%d", cfg.Jira.Issue.ProjectNameCustomField)
		extraFields = tcontainer.MarshalMap{
			customFieldName: r.Project,
		}
	}
	return jira.Issue{
		Fields: &jira.IssueFields{
			Description: cfg.Jira.Issue.Description,
			Project: jira.Project{
				Key: cfg.Jira.Issue.Project,
			},
			Type: jira.IssueType{
				Name: cfg.Jira.Issue.Type,
			},
			Labels:   labels,
			Summary:  r.IssueSummary(),
			Unknowns: extraFields,
		},
	}
}

type httpModule struct {
	jira *jira.Client
}

// handleGetRoot handles to GET requests for a basic reachability check
func (m httpModule) handleGetRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Ok")
}

// handlePostWebhook handles newreleases.io webhook post requests
func (m httpModule) handlePostWebhook(w http.ResponseWriter, r *http.Request) {
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
	existingIssuesQuery := newJiraIssueSearchQuery(cfg.Jira.Issue.Status, release.Project, cfg.Jira.Issue.ProjectNameCustomField)
	existingIssues, resp, err := m.jira.Issue.Search(existingIssuesQuery, &jira.SearchOptions{})
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
		if cfg.DryRun {
			log.Debug().
				Str("issue", i.Fields.Summary).
				Msg("Skipping creation of issue because Config.DryRun is enabled.")
			return
		}
		newIssue, resp, err := m.jira.Issue.Create(&i)
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
	if cfg.DryRun {
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
	resp, err = m.jira.Issue.UpdateIssue(oldestExistingIssue.ID, updates)
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

func projectExists(jiraClient *jira.Client) error {
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
		if project.Key == cfg.Jira.Issue.Project {
			projectExists = true
			break
		}
	}
	if !projectExists {
		return fmt.Errorf("project %v does not exist on your Jira server", cfg.Jira.Issue.Project)
	}
	return nil
}

func statusExists(jiraClient *jira.Client) error {
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
		if status.Name == cfg.Jira.Issue.Status {
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
		return fmt.Errorf("status %q does not exist on your Jira server for project %q. Available statuses: [%v]",
			cfg.Jira.Issue.Status, cfg.Jira.Issue.Project, statusSB.String())
	}
	return nil
}

func serveHTTP(jiraClient *jira.Client) error {
	m := httpModule{jira: jiraClient}
	http.HandleFunc("/webhook", m.handlePostWebhook)
	http.HandleFunc("/", m.handleGetRoot)
	log.Info().Uint16("port", cfg.HTTP.Port).Msg("Starting server.")
	return http.ListenAndServe(fmt.Sprintf(":%v", cfg.HTTP.Port), nil)
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

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

package jira

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/andygrunwald/go-jira"
	"github.com/rs/zerolog/log"
	"github.com/trivago/tgo/tcontainer"
)

type Client interface {
	ProjectMustExist(projectKey string) error
	StatusMustExist(statusName string) error
	FindIssuesForPackage(packageName string) ([]Issue, error)
	UpdateIssueSummary(issueRef IssueRef, newSummary string) error
	CreateIssue(issue Issue) (IssueRef, error)
}

type IssueRef struct {
	ID  string
	Key string
}

type Issue struct {
	ID          string
	Key         string
	Summary     string
	Description string
	Labels      []string
	ProjectKey  string
	TypeName    string

	PackageName        string
	PackageNameFieldID uint
}

func (i Issue) IssueRef() IssueRef {
	return IssueRef{
		ID:  i.ID,
		Key: i.Key,
	}
}

func newIssue(issue jira.Issue, pkgCustomFieldID uint) Issue {
	fields := util.Deref(issue.Fields, jira.IssueFields{})
	var pkgName string
	if pkgCustomFieldID != 0 {
		if str, ok := fields.Unknowns[customFieldName(pkgCustomFieldID)].(string); ok {
			pkgName = str
		}
	} else {
		// TODO: Try find pacakge name from label
	}

	return Issue{
		ID:          issue.ID,
		Key:         issue.Key,
		Summary:     fields.Summary,
		Description: fields.Description,
		Labels:      fields.Labels,

		PackageName:        pkgName,
		PackageNameFieldID: pkgCustomFieldID,
	}
}

func (i Issue) rawIssue() jira.Issue {
	labels := i.Labels
	var extraFields tcontainer.MarshalMap

	if i.PackageName != "" {
		if i.PackageNameFieldID == 0 {
			labels = append(labels, i.PackageName)
		} else {
			extraFields = tcontainer.MarshalMap{
				customFieldName(i.PackageNameFieldID): i.PackageName,
			}
		}
	}
	return jira.Issue{
		Fields: &jira.IssueFields{
			Description: i.Description,
			Project: jira.Project{
				Key: i.ProjectKey,
			},
			Type: jira.IssueType{
				Name: i.TypeName,
			},
			Labels:   i.Labels,
			Summary:  i.Summary,
			Unknowns: extraFields,
		},
	}
}

func customFieldName(fieldID uint) string {
	if fieldID == 0 {
		return ""
	}
	return fmt.Sprintf("customfield_%d", fieldID)
}

type client struct {
	cfg *config.Jira
	raw *jira.Client
}

func New(cfg *config.Jira) (Client, error) {
	var httpClient *http.Client
	tlsConfig := tls.Config{InsecureSkipVerify: cfg.SkipCertVerify}

	switch cfg.Auth.Type {
	case config.JiraAuthTypePAT:
		httpClient = (&jira.PATAuthTransport{
			Token:     cfg.Auth.Token,
			Transport: &http.Transport{TLSClientConfig: &tlsConfig},
		}).Client()
	case config.JiraAuthTypeToken:
		httpClient = (&jira.BasicAuthTransport{
			Username:  cfg.Auth.User,
			Password:  cfg.Auth.Token,
			Transport: &http.Transport{TLSClientConfig: &tlsConfig},
		}).Client()
	default:
		return nil, fmt.Errorf("invalid Jira auth type %q", cfg.Auth.Type)
	}

	httpClient.Timeout = 10 * time.Second
	jiraClient, err := jira.NewClient(httpClient, cfg.URL)
	if err != nil {
		return nil, err
	}

	return &client{
		cfg: cfg,
		raw: jiraClient,
	}, nil
}

func (c *client) ProjectMustExist(projectKey string) error {
	allProjects, response, err := c.raw.Project.GetList()
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
	for _, project := range *allProjects {
		if project.Key == projectKey {
			return nil
		}
	}
	return fmt.Errorf("project %q not found", projectKey)
}

func (c *client) StatusMustExist(statusName string) error {
	allStatuses, response, err := c.raw.Status.GetAllStatuses()
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
	for _, status := range allStatuses {
		if status.Name == statusName {
			return nil
		}
	}
	var statusNames []string
	for _, status := range allStatuses {
		statusNames = append(statusNames, status.Name)
	}
	return fmt.Errorf("status %q not found in Jira, but has: %v",
		statusName, strings.Join(statusNames, ", "))
}

func (c *client) FindIssuesForPackage(packageName string) ([]Issue, error) {
	query := newJiraIssueSearchQuery(c.cfg.Issue.Status, packageName, c.cfg.Issue.ProjectNameCustomField)
	rawIssues, resp, err := c.raw.Issue.Search(query, &jira.SearchOptions{})
	if err != nil {
		err := fmt.Errorf("searching Jira for previous issues: %w", err)
		logJiraErrResponse(resp, err)
		return nil, err
	}
	issues := make([]Issue, 0, len(rawIssues))
	for _, rawIssue := range rawIssues {
		iss := newIssue(rawIssue, c.cfg.Issue.ProjectNameCustomField)
		if iss.PackageName != packageName {
			log.Debug().
				Str("package", packageName).
				Str("issuePackage", iss.PackageName).
				Msg("Ignoring ticket because it had the wrong package name.")
			// Our search query matches substrings, so we need to filter out
			// any invalid matches
			continue
		}
		issues = append(issues, iss)
	}
	return issues, nil
}

func newJiraIssueSearchQuery(statusName, projectName string, customFieldID uint) string {
	if customFieldID == 0 {
		return fmt.Sprintf("status = %q and labels = %q ORDER BY created DESC", statusName, projectName)
	}
	// Checking label as well for backward compatibility
	return fmt.Sprintf("status = %q and (labels = %q or cf[%d] ~ %[2]q) ORDER BY created DESC",
		statusName, projectName, customFieldID)
}

func logJiraErrResponse(resp *jira.Response, err error) {
	if resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Error().Err(err).Msg("Failed Jira request, and to decode Jira response body.")
			return
		}
		var obj any
		if err := json.Unmarshal(body, &obj); err != nil {
			log.Error().Err(err).Str("body", string(body)).Msg("Failed Jira request.")
			return
		}
		log.Error().Err(err).Interface("body", obj).Msg("Failed Jira request.")
		return
	}
	log.Error().Err(err).Msg("Failed to create Jira issue.")
}

func (c *client) UpdateIssueSummary(issueRef IssueRef, newSummary string) error {
	// This seems hacky, but is taken from the official examples
	// https://github.com/andygrunwald/go-jira/blob/47d27a76e84da43f6e27e1cd0f930e6763dc79d7/examples/addlabel/main.go
	// There is also a jiraClient.Issue.Update() method, but it panics and does not provide a usage example
	type summaryUpdate struct {
		Set string `json:"set" structs:"set"`
	}
	type issueUpdate struct {
		Summary []summaryUpdate `json:"summary" structs:"summary"`
	}
	updates := map[string]any{
		"update": issueUpdate{
			Summary: []summaryUpdate{
				{Set: newSummary},
			},
		},
	}
	log.Trace().Interface("updates", updates).Msg("Updating issue.")
	resp, err := c.raw.Issue.UpdateIssue(issueRef.ID, updates)
	if err != nil {
		err := fmt.Errorf("update Jira issue: %w", err)
		logJiraErrResponse(resp, err)
		return err
	}
	log.Info().
		Str("issue", issueRef.Key).
		Str("summary", newSummary).
		Msg("Updated issue summary")
	return nil
}

func (c *client) CreateIssue(issue Issue) (IssueRef, error) {
	req := issue.rawIssue()
	created, resp, err := c.raw.Issue.Create(&req)
	if err != nil {
		err := fmt.Errorf("creating Jira issue: %w", err)
		logJiraErrResponse(resp, err)
		return IssueRef{}, err
	}
	log.Info().Str("issue", created.Key).Msg("Created issue.")
	// NOTE: Jira's "create issue" endpoint only contains the ID and Key fields
	return IssueRef{
		ID:  created.ID,
		Key: created.Key,
	}, nil
}

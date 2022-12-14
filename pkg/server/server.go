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

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/andygrunwald/go-jira"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type HTTPServer struct {
	engine *gin.Engine
	cfg    *config.Config
	jira   *jira.Client
}

func New(cfg *config.Config, jira *jira.Client) *HTTPServer {
	r := gin.New()

	r.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/health"},
			Output:    log.Logger.Level(zerolog.DebugLevel),
		}),
		gin.RecoveryWithWriter(log.Logger.Level(zerolog.ErrorLevel)),
	)

	s := &HTTPServer{
		engine: r,
		cfg:    cfg,
		jira:   jira,
	}

	r.GET("/", s.handleGetRoot)
	r.POST("/webhook", s.handlePostWebhook)

	return s
}

func (s HTTPServer) Serve() error {
	log.Info().Uint16("port", s.cfg.HTTP.Port).Msg("Starting server.")
	return s.engine.Run(fmt.Sprintf(":%v", s.cfg.HTTP.Port))
}

// handleGetRoot handles to GET requests for a basic reachability check
func (HTTPServer) handleGetRoot(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain", []byte("OK"))
}

// handlePostWebhook handles newreleases.io webhook post requests
func (s HTTPServer) handlePostWebhook(c *gin.Context) {
	// parse newreleases.io webhook
	var release Release
	if err := c.ShouldBindJSON(&release); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// look for existing update tickets
	existingIssuesQuery := newJiraIssueSearchQuery(s.cfg.Jira.Issue.Status, release.Project, s.cfg.Jira.Issue.ProjectNameCustomField)
	existingIssues, resp, err := s.jira.Issue.Search(existingIssuesQuery, &jira.SearchOptions{})
	if err != nil {
		err := fmt.Errorf("searching Jira for previous issues: %w", err)
		logJiraErrResponse(resp, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(existingIssues) == 0 {
		// no previous issues, create new jira issue
		i := release.JiraIssue(s.cfg)
		log.Trace().Interface("issue", i).Msg("Creating issue.")
		if s.cfg.DryRun {
			log.Debug().
				Str("issue", i.Fields.Summary).
				Msg("Skipping creation of issue because Config.DryRun is enabled.")
			c.Status(http.StatusNoContent)
			return
		}
		newIssue, resp, err := s.jira.Issue.Create(&i)
		if err != nil {
			err := fmt.Errorf("creating Jira issue: %w", err)
			logJiraErrResponse(resp, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		log.Info().Str("issue", newIssue.Key).Msg("Created issue.")
		c.Status(http.StatusCreated)
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
	if s.cfg.DryRun {
		log.Debug().
			Str("issue", oldestExistingIssue.Key).
			Str("summary", release.IssueSummary()).
			Msg("Skipping update of issue because Config.DryRun is enabled.")
		c.Status(http.StatusNoContent)
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
	resp, err = s.jira.Issue.UpdateIssue(oldestExistingIssue.ID, updates)
	if err != nil {
		err := fmt.Errorf("update Jira issue: %w", err)
		logJiraErrResponse(resp, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	log.Info().
		Str("issue", oldestExistingIssue.Key).
		Str("from", previousSummary).
		Str("to", release.IssueSummary()).
		Msg("Updated issue summary")
	c.Status(http.StatusNoContent)
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

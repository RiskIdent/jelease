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
	"fmt"
	"net/http"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/jira"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type HTTPServer struct {
	engine *gin.Engine
	cfg    *config.Config
	jira   jira.Client
}

func New(cfg *config.Config, jira jira.Client) *HTTPServer {
	gin.DefaultErrorWriter = log.Logger
	gin.DefaultWriter = log.Logger

	r := gin.New()

	r.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/health"},
		}),
		gin.Recovery(),
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issueRef, err := ensureJiraIssue(s.jira, release, s.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pkg, ok := s.cfg.TryFindPackage(release.Project)
	if !ok {
		log.Info().Str("project", release.Project).Msg("No package patching config was found. Skipping patching.")
		c.Status(http.StatusOK)
		// TODO: Post comment in Jira ticket.
		return
	}
	tmplCtx := patch.TemplateContext{
		Package:   release.Project,
		Version:   release.Version,
		JiraIssue: issueRef.Key,
	}
	prs, err := patch.CloneAllAndPublishPatches(s.cfg, pkg.Repos, tmplCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// TODO: Post comment in Jira ticket.
		return
	}
	if len(prs) == 0 {
		log.Warn().Str("project", release.Project).Msg("Found package config, but no repositories were patched.")
		c.Status(http.StatusOK)
		// TODO: Post comment in Jira ticket.
		return
	}
	log.Info().
		Str("project", release.Project).
		Int("count", len(prs)).
		Msg("Successfully created PRs for update.")
	c.Status(http.StatusOK)
	// TODO: Post comment in Jira ticket.
}

type newJiraIssue struct {
	jira.IssueRef
	Created bool
}

func ensureJiraIssue(j jira.Client, r Release, cfg *config.Config) (newJiraIssue, error) {
	existingIssues, err := j.FindIssuesForPackage(r.Project)
	if err != nil {
		return newJiraIssue{}, err
	}

	if len(existingIssues) == 0 {
		// no previous issues, create new jira issue
		i := r.JiraIssue(&cfg.Jira.Issue)

		log.Trace().Interface("issue", i).Msg("Creating issue.")
		if cfg.DryRun {
			log.Info().
				Str("issue", i.Key).
				Msg("Skipping creation of issue because Config.DryRun is enabled.")
			return newJiraIssue{
				IssueRef: i.IssueRef(),
				Created:  false,
			}, nil
		}
		issueRef, err := j.CreateIssue(i)
		if err != nil {
			return newJiraIssue{}, err
		}
		return newJiraIssue{
			IssueRef: issueRef,
			Created:  true,
		}, nil
	}

	// in case of duplicate issues, update the oldest (probably original) one, ignore rest as duplicates
	mostRecentIssue := existingIssues[0]
	var duplicateIssueKeys []string
	for _, issue := range existingIssues[1:] {
		duplicateIssueKeys = append(duplicateIssueKeys, issue.Key)
	}

	if len(duplicateIssueKeys) > 0 {
		log.Debug().
			Str("recent", mostRecentIssue.Key).
			Strs("duplicates", duplicateIssueKeys).
			Msg("Ignoring the duplicate issues in favor of recent issue.")
	}

	if cfg.DryRun {
		log.Debug().
			Str("issue", mostRecentIssue.Key).
			Msg("Skipping update of issue because Config.DryRun is enabled.")
		return newJiraIssue{
			IssueRef: mostRecentIssue.IssueRef(),
			Created:  false,
		}, nil
	}
	if err := j.UpdateIssueSummary(mostRecentIssue.ID, mostRecentIssue.Key, r.IssueSummary()); err != nil {
		return newJiraIssue{}, err
	}
	return newJiraIssue{
		IssueRef: mostRecentIssue.IssueRef(),
		Created:  false,
	}, nil
}

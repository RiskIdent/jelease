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
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type HTTPServer struct {
	engine *gin.Engine
	cfg    *config.Config
	jira   jira.Client
}

func New(cfg *config.Config, jira jira.Client) *HTTPServer {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingIssues, err := s.jira.FindIssuesForPackage(release.Project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(existingIssues) == 0 {
		// no previous issues, create new jira issue
		i := release.JiraIssue(&s.cfg.Jira.Issue)

		log.Trace().Interface("issue", i).Msg("Creating issue.")
		if s.cfg.DryRun {
			log.Debug().
				Str("issue", i.Key).
				Msg("Skipping creation of issue because Config.DryRun is enabled.")
			c.Status(http.StatusNoContent)
			return
		}

		if _, err := s.jira.CreateIssue(i); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Status(http.StatusCreated)
		return
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

	if s.cfg.DryRun {
		log.Debug().
			Str("issue", mostRecentIssue.Key).
			Msg("Skipping update of issue because Config.DryRun is enabled.")
		c.Status(http.StatusNoContent)
		return
	}
	if err := s.jira.UpdateIssueSummary(mostRecentIssue.ID, mostRecentIssue.Key, release.IssueSummary()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Status(http.StatusNoContent)
}

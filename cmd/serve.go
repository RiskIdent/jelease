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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/server"
	jira "github.com/andygrunwald/go-jira"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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
	s := server.New(&cfg, jiraClient)
	return s.Serve()
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

	if err = projectExists(jiraClient); err != nil {
		return nil, fmt.Errorf("check if configured project exists: %w", err)
	}
	log.Debug().Str("project", cfg.Jira.Issue.Project).Msg("Configured project found ✓")

	if err = statusExists(jiraClient); err != nil {
		return nil, fmt.Errorf("check if configured default status exists: %w", err)
	}
	log.Debug().Str("status", cfg.Jira.Issue.Status).Msg("Configured default status found ✓")

	return jiraClient, nil
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

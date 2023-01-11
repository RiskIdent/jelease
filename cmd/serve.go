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
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/RiskIdent/jelease/pkg/jira"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/RiskIdent/jelease/pkg/server"
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
	jiraClient, err := jira.New(&cfg.Jira)
	if err != nil {
		return fmt.Errorf("create jira client: %w", err)
	}

	if err := jiraClient.ProjectMustExist(cfg.Jira.Issue.Project); err != nil {
		return fmt.Errorf("check if configured project exists: %w", err)
	}
	log.Debug().Str("project", cfg.Jira.Issue.Project).Msg("Configured project found ✓")

	if err := jiraClient.StatusMustExist(cfg.Jira.Issue.Status); err != nil {
		return fmt.Errorf("check if configured default status exists: %w", err)
	}
	log.Debug().Str("status", cfg.Jira.Issue.Status).Msg("Configured default status found ✓")

	patcher, err := newTestedPatcher()
	if err != nil {
		return err
	}

	s := server.New(&cfg, jiraClient, patcher)
	return s.Serve()
}

func newTestedPatcher() (patch.Patcher, error) {
	patcher, err := patch.New(&cfg)
	if err != nil {
		return patch.Patcher{}, err
	}
	if err := patcher.TestGitHubConnection(context.TODO()); err != nil {
		return patch.Patcher{}, err
	}
	return patcher, nil
}

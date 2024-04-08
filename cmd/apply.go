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
	"fmt"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var applyFlags = struct {
	jiraIssueKey string
}{}

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:  "apply <package> <version>",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgName := args[0]
		version := args[1]
		pkg, ok := cfg.TryFindPackage(pkgName)
		if !ok {
			return fmt.Errorf("no such package found in config: %s", pkgName)
		}
		log.Info().Str("package", pkgName).Msg("Found package config")

		tmplCtx := config.TemplateContext{
			Package:   pkgName,
			Version:   version,
			JiraIssue: applyFlags.jiraIssueKey,
		}

		patcher, err := newTestedPatcher()
		if err != nil {
			return err
		}
		_, err = patcher.CloneAndPublishAll(pkg.Repos, tmplCtx)
		return err
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().StringVar(&applyFlags.jiraIssueKey, "jira.issue.key", applyFlags.jiraIssueKey, "Optional Jira ticket key used in templates, e.g PROJ-1234")
}

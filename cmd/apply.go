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

	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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

		tmplCtx := patch.TemplateContext{
			Package: pkgName,
			Version: version,
		}

		patch.CloneAllAndPublishPatches(&cfg, pkg.Repos, tmplCtx)

		log.Info().Str("package", pkgName).Msg("Done applying patches")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

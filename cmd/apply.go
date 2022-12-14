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
	"fmt"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/google/go-github/v48/github"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:  "apply <package> <version>",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgName := args[0]
		version := args[1]
		pkg, ok := tryFindPackageConfig(pkgName)
		if !ok {
			return fmt.Errorf("no such package found in config: %s", pkgName)
		}
		log.Info().Str("package", pkgName).Msg("Found package config")

		if len(pkg.Repos) == 0 {
			log.Warn().Str("package", pkgName).Msg("No repos configured for package.")
			return nil
		}

		tmplCtx := patch.TemplateContext{
			Package: pkgName,
			Version: version,
		}

		for _, pkgRepo := range pkg.Repos {
			log.Info().Str("repo", pkgRepo.URL).Msg("Patching repo")
			if err := applyRepoPatches(pkgRepo, tmplCtx); err != nil {
				return err
			}
		}

		log.Info().Str("package", pkgName).Msg("Done applying patches")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

func tryFindPackageConfig(pkgName string) (config.Package, bool) {
	for _, pkg := range cfg.Packages {
		if pkg.Name == pkgName {
			return pkg, true
		}
	}
	return config.Package{}, false
}

func applyRepoPatches(pkgRepo config.PackageRepo, tmplCtx patch.TemplateContext) error {
	g := git.Cmd{Credentials: git.Credentials{}}
	patcher, err := patch.CloneRepoForPatching(&cfg, g, pkgRepo.URL, tmplCtx)
	if err != nil {
		return err
	}
	defer patcher.Close()

	if err := patcher.ApplyManyInNewBranch(pkgRepo.Patches); err != nil {
		return err
	}

	return publishChangesUnlessDryRun(patcher)
}

func publishChangesUnlessDryRun(patcher *patch.PackagePatcher) error {
	if cfg.DryRun {
		log.Info().Msg("Dry run: skipping publishing changes.")
		return nil
	}
	gh, err := newGitHubClient()
	if err != nil {
		return fmt.Errorf("new GitHub client: %w", err)
	}

	if err := patcher.PublishChanges(gh); err != nil {
		return err
	}
	log.Info().Msg("Pushed changes to remote repository.")
	return nil
}

func newGitHubClient() (*github.Client, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GitHub.Auth.Token})
	tc := oauth2.NewClient(context.TODO(), ts)
	if cfg.GitHub.URL != nil {
		return github.NewEnterpriseClient(*cfg.GitHub.URL, "", tc)
	}
	return github.NewClient(tc), nil
}

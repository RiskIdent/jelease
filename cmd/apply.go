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
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/google/go-github/v48/github"
	"github.com/rs/zerolog"
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
	if len(pkgRepo.Patches) == 0 {
		log.Warn().
			Str("package", tmplCtx.Package).
			Str("repo", pkgRepo.URL).
			Msg("No patches configured for repository.")
		return nil
	}

	// Check this early so we don't fail right on the finish line
	repoRef, err := getGitHubRepoRef(pkgRepo.URL)
	if err != nil {
		return err
	}

	g := git.Cmd{Credentials: git.Credentials{}}
	repo, err := prepareRepo(g, pkgRepo.URL, tmplCtx)
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Warn().Err(err).Str("dir", repo.Directory()).
				Msg("Failed to clean up cloned temporary repo directory.")
		} else {
			log.Debug().Str("dir", repo.Directory()).
				Msg("Cleaned up cloned temporary repo directory.")
		}
	}()

	for _, p := range pkgRepo.Patches {
		if err := patch.Apply(repo.Directory(), p, tmplCtx); err != nil {
			return err
		}
	}

	if err := commitAndPushChanges(g, repo, tmplCtx); err != nil {
		return err
	}

	return createPR(repo, repoRef, tmplCtx)
}

func prepareRepo(g git.Git, repoURL string, tmplCtx patch.TemplateContext) (git.Repo, error) {
	dir, err := createRepoTempDirectory()
	if err != nil {
		return nil, err
	}
	repo, err := g.Clone(dir, repoURL)
	if err != nil {
		return nil, err
	}
	log.Info().
		Str("branch", repo.CurrentBranch()).
		Str("dir", repo.Directory()).
		Msg("Cloned repo.")
	branchName, err := cfg.GitHub.PR.Branch.Render(tmplCtx)
	if err != nil {
		return nil, fmt.Errorf("template branch name: %w", err)
	}
	if err := repo.CheckoutNewBranch(branchName); err != nil {
		return nil, err
	}
	log.Info().Str("branch", repo.CurrentBranch()).Str("mainBranch", repo.MainBranch()).Msg("Checked out new branch.")
	return repo, nil
}

func createRepoTempDirectory() (string, error) {
	parentDir := filepath.Join(util.Deref(cfg.GitHub.TempDir, os.TempDir()), "jelease-cloned-repos")
	if err := os.MkdirAll(parentDir, 0700); err != nil {
		return "", err
	}
	return os.MkdirTemp(parentDir, "jelease-repo-*")
}

func commitAndPushChanges(g git.Git, repo git.Repo, tmplCtx patch.TemplateContext) error {
	logDiff(repo)

	if err := repo.StageChanges(); err != nil {
		return err
	}
	log.Debug().Msg("Staged changes.")

	commitMsg, err := cfg.GitHub.PR.Commit.Render(tmplCtx)
	if err != nil {
		return fmt.Errorf("template commit message: %w", err)
	}
	commit, err := repo.CreateCommit(commitMsg)
	if err != nil {
		return err
	}
	log.Debug().
		Str("hash", commit.AbbrHash).
		Str("subject", commit.Subject).
		Msg("Created commit.")

	if cfg.DryRun {
		log.Info().Msg("Dry run: skipping pushing changes.")
		return nil
	}

	if err := repo.PushChanges(); err != nil {
		return err
	}
	log.Info().Msg("Pushed changes to remote repository.")
	return nil
}

func logDiff(repo git.Repo) {
	if log.Logger.GetLevel() > zerolog.DebugLevel {
		return
	}
	diff, err := repo.DiffChanges()
	if err != nil {
		log.Warn().Err(err).Msg("Failed diffing changes. Trying to continue anyways.")
		return
	}
	if cfg.Log.Format == config.LogFormatPretty {
		diff = git.ColorizeDiff(diff)
	}
	log.Debug().Msgf("Diff:\n%s", diff)
}

func createPR(repo git.Repo, repoRef GitHubRepoRef, tmplCtx patch.TemplateContext) error {
	gh, err := newGitHubClient()
	if err != nil {
		return fmt.Errorf("new GitHub client: %w", err)
	}

	title, err := cfg.GitHub.PR.Title.Render(tmplCtx)
	if err != nil {
		return fmt.Errorf("template PR title: %w", err)
	}
	description, err := cfg.GitHub.PR.Description.Render(tmplCtx)
	if err != nil {
		return fmt.Errorf("template PR description: %w", err)
	}

	if cfg.DryRun {
		log.Info().Msg("Dry run: skipping creating GitHub pull request.")
		return nil
	}

	pr, _, err := gh.PullRequests.Create(context.TODO(), repoRef.Owner, repoRef.Repo, &github.NewPullRequest{
		Title:               &title,
		Body:                &description,
		Head:                util.Ref(repo.CurrentBranch()),
		Base:                util.Ref(repo.MainBranch()),
		MaintainerCanModify: util.Ref(true),
	})
	if err != nil {
		return fmt.Errorf("create GitHub PR: %w", err)
	}
	log.Info().
		Int("pr", util.Deref(pr.Number, -1)).
		Str("url", util.Deref(pr.HTMLURL, "")).
		Msg("GitHub PR created.")
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

type GitHubRepoRef struct {
	Owner string
	Repo  string
}

func getGitHubRepoRef(remote string) (GitHubRepoRef, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return GitHubRepoRef{}, err
	}
	u.User = nil
	path := u.Path
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) < 2 {
		return GitHubRepoRef{}, fmt.Errorf("expected https://host/OWNER/REPO in URL, got: %s", u.String())
	}
	return GitHubRepoRef{
		Owner: segments[0],
		Repo:  strings.TrimSuffix(segments[1], ".git"),
	}, nil
}

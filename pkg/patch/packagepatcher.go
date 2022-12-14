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

package patch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/google/go-github/v48/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func CloneRepoForPatching(cfg *config.Config, g git.Git, remote string, tmplCtx TemplateContext) (*PackagePatcher, error) {
	// Check this early so we don't fail right on the finish line
	ghRef, err := ParseGitHubRepoRef(remote)
	if err != nil {
		return nil, err
	}
	repo, err := cloneRepoTemp(g, util.Deref(cfg.GitHub.TempDir, os.TempDir()), remote, tmplCtx)
	if err != nil {
		return nil, err
	}
	return &PackagePatcher{
		ghRef:   ghRef,
		remote:  remote,
		repo:    repo,
		cfg:     cfg,
		tmplCtx: tmplCtx,
	}, nil
}

func cloneRepoTemp(g git.Git, tempDir, repoURL string, tmplCtx TemplateContext) (git.Repo, error) {
	targetDir := filepath.Join(tempDir, "jelease-cloned-repos", "jelease-repo-*")
	repo, err := git.CloneTemp(g, targetDir, repoURL)
	if err != nil {
		return nil, err
	}
	log.Info().
		Str("branch", repo.CurrentBranch()).
		Str("dir", repo.Directory()).
		Msg("Cloned repo.")
	return repo, nil
}

type PackagePatcher struct {
	ghRef   GitHubRepoRef
	remote  string
	repo    git.Repo
	cfg     *config.Config
	tmplCtx TemplateContext
}

func (p *PackagePatcher) Close() error {
	err := p.repo.Close()
	if err != nil {
		log.Warn().Err(err).Str("dir", p.repo.Directory()).
			Msg("Failed to clean up cloned temporary repo directory.")
	} else {
		log.Debug().Str("dir", p.repo.Directory()).
			Msg("Cleaned up cloned temporary repo directory.")
	}
	return err
}

func (p *PackagePatcher) ApplyManyInNewBranch(patches []config.PackageRepoPatch) error {
	if len(patches) == 0 {
		log.Warn().
			Str("package", p.tmplCtx.Package).
			Str("repo", p.remote).
			Msg("No patches configured for repository.")
		return nil
	}

	if err := p.ApplyMany(patches); err != nil {
		return fmt.Errorf("template branch name: %w", err)
	}

	if err := p.repo.StageChanges(); err != nil {
		return err
	}
	log.Debug().Msg("Staged changes.")

	commitMsg, err := p.cfg.GitHub.PR.Commit.Render(p.tmplCtx)
	if err != nil {
		return fmt.Errorf("template commit message: %w", err)
	}
	commit, err := p.repo.CreateCommit(commitMsg)
	if err != nil {
		return err
	}
	log.Debug().
		Str("hash", commit.AbbrHash).
		Str("subject", commit.Subject).
		Msg("Created commit.")
	return nil
}

func (p *PackagePatcher) ApplyMany(patches []config.PackageRepoPatch) error {
	branchName, err := p.cfg.GitHub.PR.Branch.Render(p.tmplCtx)
	if err != nil {
		return fmt.Errorf("template branch name: %w", err)
	}
	if err := p.repo.CheckoutNewBranch(branchName); err != nil {
		return err
	}
	log.Debug().
		Str("branch", p.repo.CurrentBranch()).
		Str("base", p.repo.MainBranch()).
		Msg("Checked out new branch.")
	if err := ApplyMany(p.repo.Directory(), patches, p.tmplCtx); err != nil {
		return err
	}

	p.logDiff()
	return nil
}

func (p *PackagePatcher) PublishChanges(gh *github.Client) error {
	if err := p.repo.PushChanges(); err != nil {
		return err
	}
	log.Info().Str("branch", p.repo.CurrentBranch()).
		Msg("Pushed changes to remote repository.")

	title, err := p.cfg.GitHub.PR.Title.Render(p.tmplCtx)
	if err != nil {
		return fmt.Errorf("template PR title: %w", err)
	}
	description, err := p.cfg.GitHub.PR.Description.Render(p.tmplCtx)
	if err != nil {
		return fmt.Errorf("template PR description: %w", err)
	}

	pr, _, err := gh.PullRequests.Create(context.TODO(), p.ghRef.Owner, p.ghRef.Repo, &github.NewPullRequest{
		Title:               &title,
		Body:                &description,
		Head:                util.Ref(p.repo.CurrentBranch()),
		Base:                util.Ref(p.repo.MainBranch()),
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

func (p *PackagePatcher) logDiff() {
	if log.Logger.GetLevel() > zerolog.DebugLevel {
		return
	}
	diff, err := p.repo.DiffChanges()
	if err != nil {
		log.Warn().Err(err).Msg("Failed diffing changes. Trying to continue anyways.")
		return
	}
	if p.cfg.Log.Format == config.LogFormatPretty {
		diff = git.ColorizeDiff(diff)
	}
	log.Debug().Msgf("Diff:\n%s", diff)
}

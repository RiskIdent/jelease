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

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/RiskIdent/jelease/pkg/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Repo is an object to manage patches for a single repository.
// It is obtained from the [Patcher] object.
type Repo struct {
	gh      github.Client
	ghRef   github.RepoRef
	remote  string
	repo    git.Repo
	cfg     *config.Config
	tmplCtx config.TemplateContext
}

// Close cleans up the Git repository by removing the entire directory.
func (p *Repo) Close() error {
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

// ApplyManyAndCommit uses [ApplyManyInNewBranch] to apply a series of patches
// in a new Git branch, followed by creating a Git commit.
func (p *Repo) ApplyManyAndCommit(patches []config.PackageRepoPatch) (git.Commit, error) {
	if len(patches) == 0 {
		log.Warn().
			Str("package", p.tmplCtx.Package).
			Str("repo", p.remote).
			Msg("No patches configured for repository.")
		return git.Commit{}, ErrNoPatches
	}

	if err := p.ApplyManyInNewBranch(patches); err != nil {
		return git.Commit{}, fmt.Errorf("template branch name: %w", err)
	}

	if err := p.repo.StageChanges(); err != nil {
		return git.Commit{}, err
	}
	log.Debug().Msg("Staged changes.")

	commitMsg, err := p.cfg.GitHub.PR.Commit.Render(p.tmplCtx)
	if err != nil {
		return git.Commit{}, fmt.Errorf("template commit message: %w", err)
	}
	commit, err := p.repo.CreateCommit(commitMsg)
	if err != nil {
		return git.Commit{}, err
	}
	log.Debug().
		Str("hash", commit.AbbrHash).
		Str("subject", commit.Subject).
		Msg("Created commit.")
	p.logDiff(commit.Diff)
	return commit, nil
}

// ApplyManyInNewBranch creates a new Git branch and then applies multiple
// patches in series using [ApplyMany].
func (p *Repo) ApplyManyInNewBranch(patches []config.PackageRepoPatch) error {
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

	return nil
}

// PublishChangesUnlessDryRun calls [PublishChanges], unless dry-run is set
// in the configs.
//
// If dry-run is enabled, then this function templates all PR fields
// (title, description, etc), and then returns it as-is without pushing anything
// to the Git remote.
func (p *Repo) PublishChangesUnlessDryRun(commit git.Commit) (github.PullRequest, error) {
	if p.cfg.DryRun {
		log.Info().Msg("Dry run: skipping publishing changes.")
		newPR, err := p.TemplateNewPullRequest(commit)
		if err != nil {
			return github.PullRequest{}, err
		}
		return github.PullRequest{
			RepoRef:     newPR.RepoRef,
			Title:       newPR.Title,
			Description: newPR.Description,
			Head:        newPR.Head,
			Base:        newPR.Base,
			Commit:      newPR.Commit,
		}, nil
	}
	pr, err := p.PublishChanges(commit)
	if err != nil {
		return github.PullRequest{}, err
	}
	log.Info().Msg("Pushed changes to remote repository.")
	return pr, nil
}

// PublishChanges will push the current Git branch to the remote, and then
// create a GitHub pull request.
func (p *Repo) PublishChanges(commit git.Commit) (github.PullRequest, error) {
	if err := p.repo.PushChanges(); err != nil {
		return github.PullRequest{}, err
	}
	log.Info().Str("branch", p.repo.CurrentBranch()).
		Msg("Pushed changes to remote repository.")

	newPR, err := p.TemplateNewPullRequest(commit)
	if err != nil {
		return github.PullRequest{}, err
	}

	pr, err := p.gh.CreatePullRequest(context.TODO(), newPR)
	if err != nil {
		return github.PullRequest{}, fmt.Errorf("create GitHub PR: %w", err)
	}
	log.Info().
		Str("url", pr.URL).
		Msg("GitHub PR created.")
	return pr, nil
}

// TemplateNewPullRequest will template using [text/template] the pull request
// fields (title, description, etc), based on what's set in the config.
func (p *Repo) TemplateNewPullRequest(commit git.Commit) (github.NewPullRequest, error) {
	title, err := p.cfg.GitHub.PR.Title.Render(p.tmplCtx)
	if err != nil {
		return github.NewPullRequest{}, fmt.Errorf("template PR title: %w", err)
	}
	description, err := p.cfg.GitHub.PR.Description.Render(p.tmplCtx)
	if err != nil {
		return github.NewPullRequest{}, fmt.Errorf("template PR description: %w", err)
	}

	return github.NewPullRequest{
		RepoRef:     p.ghRef,
		Title:       title,
		Description: description,
		Head:        p.repo.CurrentBranch(),
		Base:        p.repo.MainBranch(),
		Commit:      commit,
	}, nil
}

// logDiff sends a log message with a commit diff. Will optionally colorize it
// if log format is set to "pretty".
func (p *Repo) logDiff(diff string) {
	if log.Logger.GetLevel() > zerolog.DebugLevel {
		return
	}
	if p.cfg.Log.Format == config.LogFormatPretty {
		diff = git.ColorizeDiff(diff)
	}
	log.Debug().Msgf("Diff:\n%s", diff)
}

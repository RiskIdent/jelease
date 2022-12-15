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
	"fmt"
	"os"
	"path/filepath"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/RiskIdent/jelease/pkg/github"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func CloneAllAndPublishPatches(cfg *config.Config, pkgRepos []config.PackageRepo, tmplCtx TemplateContext) ([]github.PullRequest, error) {
	if len(pkgRepos) == 0 {
		log.Warn().Str("package", tmplCtx.Package).Msg("No repos configured for package.")
		return nil, nil
	}
	g := git.Cmd{
		Credentials: git.Credentials{Username: "git", Password: cfg.GitHub.Auth.Token},
		Committer: git.Committer{
			Name:  util.Deref(cfg.GitHub.PR.Committer.Name, ""),
			Email: util.Deref(cfg.GitHub.PR.Committer.Email, ""),
		},
	}
	gh, err := github.New(&cfg.GitHub)
	if err != nil {
		return nil, err
	}

	var prs []github.PullRequest
	for _, pkgRepo := range pkgRepos {
		log.Info().Str("repo", pkgRepo.URL).Msg("Patching repo")
		pr, err := CloneRepoAndPublishPatches(cfg, g, gh, pkgRepo, tmplCtx)
		if err != nil {
			return prs, err
		}
		prs = append(prs, pr)
	}

	log.Info().Str("package", tmplCtx.Package).Msg("Done applying patches")
	return prs, nil
}

func CloneRepoAndPublishPatches(cfg *config.Config, g git.Git, gh github.Client, pkgRepo config.PackageRepo, tmplCtx TemplateContext) (github.PullRequest, error) {
	patcher, err := CloneRepoForPatching(cfg, g, gh, pkgRepo.URL, tmplCtx)
	if err != nil {
		return github.PullRequest{}, err
	}
	defer patcher.Close()

	if err := patcher.ApplyManyAndCommit(pkgRepo.Patches); err != nil {
		return github.PullRequest{}, err
	}

	return patcher.PublishChangesUnlessDryRun()
}

func CloneRepoForPatching(cfg *config.Config, g git.Git, gh github.Client, remote string, tmplCtx TemplateContext) (*PackagePatcher, error) {
	// Check this early so we don't fail right on the finish line
	ghRef, err := github.ParseRepoRef(remote)
	if err != nil {
		return nil, err
	}
	repo, err := cloneRepoTemp(g, util.Deref(cfg.GitHub.TempDir, os.TempDir()), remote, tmplCtx)
	if err != nil {
		return nil, err
	}
	return &PackagePatcher{
		gh:      gh,
		ghRef:   ghRef,
		remote:  remote,
		repo:    repo,
		cfg:     cfg,
		tmplCtx: tmplCtx,
	}, nil
}

func cloneRepoTemp(g git.Git, tempDir, remote string, tmplCtx TemplateContext) (git.Repo, error) {
	targetDir := filepath.Join(tempDir, "jelease-cloned-repos", "jelease-repo-*")
	repo, err := git.CloneTemp(g, targetDir, remote)
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
	gh      github.Client
	ghRef   github.RepoRef
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

func (p *PackagePatcher) ApplyManyAndCommit(patches []config.PackageRepoPatch) error {
	if len(patches) == 0 {
		log.Warn().
			Str("package", p.tmplCtx.Package).
			Str("repo", p.remote).
			Msg("No patches configured for repository.")
		return nil
	}

	if err := p.ApplyManyInNewBranch(patches); err != nil {
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

func (p *PackagePatcher) ApplyManyInNewBranch(patches []config.PackageRepoPatch) error {
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

func (p *PackagePatcher) PublishChanges() (github.PullRequest, error) {
	if err := p.repo.PushChanges(); err != nil {
		return github.PullRequest{}, err
	}
	log.Info().Str("branch", p.repo.CurrentBranch()).
		Msg("Pushed changes to remote repository.")

	title, err := p.cfg.GitHub.PR.Title.Render(p.tmplCtx)
	if err != nil {
		return github.PullRequest{}, fmt.Errorf("template PR title: %w", err)
	}
	description, err := p.cfg.GitHub.PR.Description.Render(p.tmplCtx)
	if err != nil {
		return github.PullRequest{}, fmt.Errorf("template PR description: %w", err)
	}

	pr, err := p.gh.CreatePullRequest(github.NewPullRequest{
		RepoRef:     p.ghRef,
		Title:       title,
		Description: description,
		Head:        p.repo.CurrentBranch(),
		Base:        p.repo.MainBranch(),
	})
	if err != nil {
		return github.PullRequest{}, fmt.Errorf("create GitHub PR: %w", err)
	}
	log.Info().
		Int("pr", pr.Number).
		Str("url", pr.URL).
		Msg("GitHub PR created.")
	return pr, nil
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

func (p *PackagePatcher) PublishChangesUnlessDryRun() (github.PullRequest, error) {
	if p.cfg.DryRun {
		log.Info().Msg("Dry run: skipping publishing changes.")
		return github.PullRequest{}, nil
	}
	pr, err := p.PublishChanges()
	if err != nil {
		return github.PullRequest{}, err
	}
	log.Info().Msg("Pushed changes to remote repository.")
	return pr, nil
}

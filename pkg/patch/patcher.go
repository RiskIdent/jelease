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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/RiskIdent/jelease/pkg/github"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/rs/zerolog/log"
)

var (
	ErrNoPatches = errors.New("no patches configured for repository")
)

type Patcher struct {
	cfg *config.Config
	gh  github.Client
}

func New(cfg *config.Config) (Patcher, error) {
	gh, err := github.New(&cfg.GitHub)
	if err != nil {
		return Patcher{}, err
	}
	return Patcher{
		cfg: cfg,
		gh:  gh,
	}, nil
}

func (p Patcher) TestGitHubConnection(ctx context.Context) error {
	if err := p.gh.TestConnection(context.TODO()); err != nil {
		return fmt.Errorf("test GitHub connection: %w", err)
	}
	return nil
}

func (p Patcher) CloneAndPublishAll(pkgRepos []config.PackageRepo, tmplCtx TemplateContext) ([]github.PullRequest, error) {
	if len(pkgRepos) == 0 {
		log.Warn().Str("package", tmplCtx.Package).Msg("No repos configured for package.")
		return nil, nil
	}

	var prs []github.PullRequest
	for _, pkgRepo := range pkgRepos {
		log.Info().Str("repo", pkgRepo.URL).Msg("Patching repo")
		pr, err := p.CloneAndPublishRepo(pkgRepo, tmplCtx)
		if errors.Is(err, ErrNoPatches) {
			continue
		}
		if err != nil {
			return prs, err
		}
		prs = append(prs, pr)
	}

	log.Info().Str("package", tmplCtx.Package).Msg("Done applying patches")
	return prs, nil
}

func (p Patcher) CloneAndPublishRepo(pkgRepo config.PackageRepo, tmplCtx TemplateContext) (github.PullRequest, error) {
	pkgPatcher, err := p.CloneRepo(pkgRepo.URL, tmplCtx)
	if err != nil {
		return github.PullRequest{}, err
	}
	defer pkgPatcher.Close()

	if err := pkgPatcher.ApplyManyAndCommit(pkgRepo.Patches); err != nil {
		return github.PullRequest{}, err
	}

	return pkgPatcher.PublishChangesUnlessDryRun()
}

func (p Patcher) CloneRepo(remote string, tmplCtx TemplateContext) (*Repo, error) {
	// Check this early so we don't fail right on the finish line
	ghRef, err := github.ParseRepoRef(remote)
	if err != nil {
		return nil, err
	}
	gitCred, err := p.gh.GitCredentialsForRepo(context.TODO(), ghRef)
	if err != nil {
		return nil, err
	}
	g := git.Cmd{
		Credentials: gitCred,
		Committer: git.Committer{
			Name:  util.Deref(p.cfg.GitHub.PR.Committer.Name, ""),
			Email: util.Deref(p.cfg.GitHub.PR.Committer.Email, ""),
		},
	}
	repo, err := cloneRepoTemp(g, util.Deref(p.cfg.GitHub.TempDir, os.TempDir()), remote, tmplCtx)
	if err != nil {
		return nil, err
	}
	return &Repo{
		gh:      p.gh,
		ghRef:   ghRef,
		remote:  remote,
		repo:    repo,
		cfg:     p.cfg,
		tmplCtx: tmplCtx,
	}, nil
}

func cloneRepoTemp(g git.Git, tempDir, remote string, tmplCtx TemplateContext) (git.Repo, error) {
	targetDir := filepath.Join(tempDir, "jelease-cloned-repos", "jelease-repo-*")
	repo, err := git.CloneTemp(g, targetDir, remote)
	if err != nil {
		return nil, err
	}
	log.Debug().
		Str("branch", repo.CurrentBranch()).
		Str("dir", repo.Directory()).
		Msg("Cloned repo.")
	return repo, nil
}

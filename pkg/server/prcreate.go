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

package server

import (
	"fmt"
	"net/http"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/github"
	"github.com/RiskIdent/jelease/pkg/jira"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CreatePRContext is the [html/template] context used when templating the
// "create PR" page.
type CreatePRContext struct {
	Config  *config.Config
	Package config.Package

	Version           string
	JiraIssue         string
	JiraCreateComment bool
	DryRun            bool
	IsPost            bool
	PullRequests      []github.PullRequest
	Error             error
}

// CreatePRRequest is the query or form data pushed by the web.
type CreatePRRequest struct {
	Version           string `form:"version"`
	JiraIssue         string `form:"jiraIssue"`
	JiraCreateComment bool   `form:"jiraCreateComment"`
	PRCreate          bool   `form:"prCreate"`
}

func (s HTTPServer) bindCreatePRContext(c *gin.Context) (CreatePRContext, bool) {
	pkgName := c.Param("package")
	pkg, ok := s.cfg.TryFindNormalizedPackage(pkgName)
	if !ok {
		c.HTML(http.StatusOK, "404", map[string]any{
			"Config": s.cfg,
			"Alert":  fmt.Sprintf("Package %q not found.", pkgName),
		})
		return CreatePRContext{}, false
	}
	var input CreatePRRequest
	err := c.ShouldBind(&input)
	model := CreatePRContext{
		Config:            s.cfg,
		Package: pkg,
		Version:           input.Version,
		JiraIssue:         input.JiraIssue,
		JiraCreateComment: input.JiraCreateComment && !s.cfg.DryRun,
		DryRun:            !input.PRCreate || s.cfg.DryRun,
		IsPost:            c.Request.Method == http.MethodPost,
	}
	if err != nil {
		model.Error = err
		c.HTML(http.StatusBadRequest, "package-create-pr", model)
		return model, false
	}

	if s.cfg.DryRun || !input.PRCreate {
		input.PRCreate = false
		input.JiraCreateComment = false
	}

	return model, true
}

// handleGetPRCreate is the handler for:
//
//	GET /packages/:package/create-pr
func (s HTTPServer) handleGetPRCreate(c *gin.Context) {
	model, ok := s.bindCreatePRContext(c)
	if !ok {
		return
	}
	c.HTML(http.StatusOK, "package-create-pr", model)
}

// handlePostPRCreate is the handler for:
//
//	POST /packages/:package/create-pr
func (s HTTPServer) handlePostPRCreate(c *gin.Context) {
	model, ok := s.bindCreatePRContext(c)
	if !ok {
		return
	}

	var issueRef jira.IssueRef
	if model.JiraCreateComment || model.JiraIssue != "" {
		issue, err := s.jira.FindIssueForKey(model.JiraIssue)
		if err != nil {
			model.Error = err
			c.HTML(http.StatusOK, "package-create-pr", model)
			return
		}
		issueRef = issue.IssueRef()
	}

	cfgClone := *s.cfg
	patcherClone := s.patcher.CloneWithConfig(&cfgClone)

	tmplCtx2 := patch.TemplateContext{
		Package:   model.Package.Name,
		Version:   model.Version,
		JiraIssue: model.JiraIssue,
	}
	prs, err := patcherClone.CloneAndPublishAll(model.Package.Repos, tmplCtx2)

	if model.JiraCreateComment {
		if len(prs) == 0 {
			log.Warn().Str("project", model.Package.Name).Msg("Found package config, but no repositories were patched.")
			createTemplatedComment(s.jira, issueRef, s.cfg.Jira.Issue.Comments.NoPatches, model)
		} else {
			log.Info().
				Str("project", model.Package.Name).
				Int("count", len(prs)).
				Msg("Successfully created PRs for update.")

			createTemplatedComment(s.jira, issueRef, s.cfg.Jira.Issue.Comments.PRCreated, TemplateContextPullRequests{
				TemplateContext: tmplCtx2,
				PullRequests:    prs,
			})
		}
	}

	model.PullRequests = prs
	model.Error = err
	c.HTML(http.StatusOK, "package-create-pr", model)
}

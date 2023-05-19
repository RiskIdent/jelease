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
	Executed          bool
	PullRequests      []github.PullRequest
	Error             error
}

func (s HTTPServer) handleGetPRCreate(c *gin.Context) {
	version := c.Query("version")
	jiraIssue := c.Query("jiraIssue")
	jiraCreateComment := parseQueryBool(c, "jiraCreateComment")
	prCreate := parseQueryBool(c, "prCreate")

	pkgName := c.Param("package")
	pkg, ok := s.cfg.TryFindNormalizedPackage(pkgName)
	if !ok {
		c.HTML(http.StatusOK, "404", map[string]any{
			"Config": s.cfg,
			"Alert":  fmt.Sprintf("Package %q not found.", pkgName),
		})
		return
	}
	c.HTML(http.StatusOK, "package-create-pr", CreatePRContext{
		Config:            s.cfg,
		Package:           pkg,
		Version:           version,
		JiraIssue:         jiraIssue,
		JiraCreateComment: jiraCreateComment,
		DryRun:            !prCreate,
		Executed:          false,
	})
}

// CreatePRRequest is the query or form data pushed by the web.
type CreatePRRequest struct {
	Version           string `form:"version"`
	JiraIssue         string `form:"jiraIssue"`
	JiraCreateComment bool   `form:"jiraCreateComment"`
	PRCreate          bool   `form:"prCreate"`
}

func (s HTTPServer) handlePostPRCreate(c *gin.Context) {
	var input CreatePRRequest
	if err := c.ShouldBind(&input); err != nil {
		// TODO: handle
		return
	}

	pkgName := c.Param("package")
	pkg, ok := s.cfg.TryFindNormalizedPackage(pkgName)
	if !ok {
		c.HTML(http.StatusOK, "package-404", map[string]any{
			"Config":      s.cfg,
			"PackageName": pkgName,
		})
		return
	}

	if s.cfg.DryRun || !input.PRCreate {
		input.PRCreate = false
		input.JiraCreateComment = false
	}

	var issueRef jira.IssueRef
	if input.JiraCreateComment || input.JiraIssue != "" {
		issue, err := s.jira.FindIssueForKey(input.JiraIssue)
		if err != nil {
			c.HTML(http.StatusOK, "package-create-pr", map[string]any{
				"Config":            s.cfg,
				"Package":           pkg,
				"Version":           input.Version,
				"JiraIssue":         input.JiraIssue,
				"JiraCreateComment": input.JiraCreateComment,
				"DryRun":            !input.PRCreate,
				"Executed":          true,

				"PullRequests": []github.PullRequest{},
				"Error":        err,
			})
			return
		}
		issueRef = issue.IssueRef()
	}

	cfgClone := *s.cfg
	cfgClone.DryRun = !input.PRCreate
	patcherClone := s.patcher.CloneWithConfig(&cfgClone)

	tmplCtx := patch.TemplateContext{
		Package:   pkg.Name,
		Version:   input.Version,
		JiraIssue: input.JiraIssue,
	}
	prs, err := patcherClone.CloneAndPublishAll(pkg.Repos, tmplCtx)

	if input.JiraCreateComment {
		if len(prs) == 0 {
			log.Warn().Str("project", pkgName).Msg("Found package config, but no repositories were patched.")
			createTemplatedComment(s.jira, issueRef, s.cfg.Jira.Issue.Comments.NoPatches, tmplCtx)
		} else {
			log.Info().
				Str("project", pkgName).
				Int("count", len(prs)).
				Msg("Successfully created PRs for update.")

			createTemplatedComment(s.jira, issueRef, s.cfg.Jira.Issue.Comments.PRCreated, TemplateContextPullRequests{
				TemplateContext: tmplCtx,
				PullRequests:    prs,
			})
		}
	}

	c.HTML(http.StatusOK, "package-create-pr", CreatePRContext{
		Config:            s.cfg,
		Package:           pkg,
		Version:           input.Version,
		JiraIssue:         input.JiraIssue,
		JiraCreateComment: input.JiraCreateComment,
		DryRun:            !input.PRCreate,
		Executed:          true,

		PullRequests: prs,
		Error:        err,
	})
}

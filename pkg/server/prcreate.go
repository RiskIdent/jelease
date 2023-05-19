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
	"net/url"
	"strings"

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

	Version      string
	JiraIssue    string
	DryRun       bool
	IsPost       bool
	PullRequests []github.PullRequest
	Error        error
}

// CreatePRRequest is the query or form data pushed by the web.
type CreatePRRequest struct {
	Version   string `form:"version"`
	JiraIssue string `form:"jiraIssue"`
	PRCreate  bool   `form:"prCreate"`
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
		Config:    s.cfg,
		Package:   pkg,
		Version:   input.Version,
		JiraIssue: input.JiraIssue,
		DryRun:    !input.PRCreate || s.cfg.DryRun,
		IsPost:    c.Request.Method == http.MethodPost,
	}
	if err != nil {
		model.Error = err
		c.HTML(http.StatusBadRequest, "package-create-pr", model)
		return model, false
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
	if model.JiraIssue != "" {
		issue, err := s.jira.FindIssueForKey(model.JiraIssue)
		if err != nil {
			model.Error = err
			c.HTML(http.StatusOK, "package-create-pr", model)
			return
		}
		issueRef = issue.IssueRef()
	}

	cfgClone := *s.cfg
	cfgClone.DryRun = model.DryRun
	patcherClone := s.patcher.CloneWithConfig(&cfgClone)

	tmplCtx := patch.TemplateContext{
		Package:   model.Package.Name,
		Version:   model.Version,
		JiraIssue: model.JiraIssue,
	}
	prs, err := patcherClone.CloneAndPublishAll(model.Package.Repos, tmplCtx)
	if err != nil {
		log.Error().Err(err).Str("project", model.Package.Name).Msg("Failed creating patches.")
	}

	if model.JiraIssue != "" && !model.DryRun && err == nil {
		createDynamicComment(s.jira, issueRef, prs, model.Package.Name, &s.cfg.Jira.Issue.Comments, tmplCtx)
	}

	model.PullRequests = prs
	model.Error = err
	c.HTML(http.StatusOK, "package-create-pr", model)
}

func createDeferredCreationURL(publicURL *url.URL, pkgName string, req CreatePRRequest) *url.URL {
	u := *publicURL
	u.Path = fmt.Sprintf("%s/packages/%s/create-pr",
		strings.TrimSuffix(u.Path, "/"),
		url.PathEscape(config.NormalizePackageName(pkgName)),
	)
	values := url.Values{}
	if req.Version != "" {
		values.Set("version", req.Version)
	}
	if req.PRCreate {
		values.Set("prCreate", "true")
	}
	if req.JiraIssue != "" {
		values.Set("jiraIssue", req.JiraIssue)
	}
	u.RawQuery = values.Encode()
	u.Fragment = ""
	return &u
}

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
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/github"
	"github.com/RiskIdent/jelease/pkg/jira"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/RiskIdent/jelease/pkg/templatefuncs"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type HTTPServer struct {
	engine  *gin.Engine
	cfg     *config.Config
	jira    jira.Client
	patcher patch.Patcher
}

func New(cfg *config.Config, j jira.Client, patcher patch.Patcher, htmlTemplates fs.FS, staticFiles fs.FS) *HTTPServer {
	gin.DefaultErrorWriter = ginLogger{defaultLevel: zerolog.ErrorLevel}
	gin.DefaultWriter = ginLogger{defaultLevel: zerolog.InfoLevel}

	r := gin.New()
	r.HandleMethodNotAllowed = true

	r.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/"},
		}),
		gin.Recovery(),
	)

	s := &HTTPServer{
		engine:  r,
		cfg:     cfg,
		jira:    j,
		patcher: patcher,
	}

	ren := multitemplate.New()
	r.HTMLRender = ren
	defaultTemplateObj := map[string]any{"Config": s.cfg.Censored()}

	addHTMLFromFS(ren, htmlTemplates, "index", "layout.html", "index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", defaultTemplateObj)
	})

	addHTMLFromFS(ren, htmlTemplates, "config", "layout.html", "config.html")
	r.GET("/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "config", defaultTemplateObj)
	})

	addHTMLFromFS(ren, htmlTemplates, "package-list", "layout.html", "packages/index.html")
	r.GET("/packages", func(c *gin.Context) {
		c.HTML(http.StatusOK, "package-list", defaultTemplateObj)
	})

	addHTMLFromFS(ren, htmlTemplates, "package-item", "layout.html", "packages/package.html")
	r.GET("/packages/:package", func(c *gin.Context) {
		pkgName := c.Param("package")
		pkg, ok := s.cfg.TryFindNormalizedPackage(pkgName)
		if !ok {
			c.HTML(http.StatusOK, "404", map[string]any{
				"Config": s.cfg,
				"Alert":  fmt.Sprintf("Package %q not found.", pkgName),
			})
			return
		}
		c.HTML(http.StatusOK, "package-item", map[string]any{
			"Config":  s.cfg,
			"Package": pkg,
		})
	})

	addHTMLFromFS(ren, htmlTemplates, "package-create-pr", "layout.html", "packages/create-pr.html")
	r.GET("/packages/:package/create-pr", s.handleGetPRCreate)
	r.POST("/packages/:package/create-pr", s.handlePostPRCreate)

	addHTMLFromFS(ren, htmlTemplates, "404", "layout.html", "404.html")
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/webhook") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			return
		}
		c.HTML(http.StatusNotFound, "404", defaultTemplateObj)
	})

	addHTMLFromFS(ren, htmlTemplates, "405", "layout.html", "405.html")
	r.NoMethod(func(c *gin.Context) {
		var methodsAllowed []string
		for _, route := range r.Routes() {
			if route.Path == c.Request.URL.Path {
				methodsAllowed = append(methodsAllowed, route.Method)
			}
		}
		if strings.HasPrefix(c.Request.URL.Path, "/webhook") {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error":   "Method not allowed",
				"methods": methodsAllowed,
			})
			return
		}
		c.HTML(http.StatusNotFound, "405", defaultTemplateObj)
	})

	r.POST("/webhook", s.handlePostWebhook)

	httpFS := http.FS(staticFiles)
	fs.WalkDir(staticFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".license") {
			return nil
		}
		r.StaticFileFS(path, path, httpFS)
		return nil
	})

	return s
}

func parseQueryBool(c *gin.Context, name string) bool {
	value, ok := c.GetQuery(name)
	if !ok {
		return false
	}
	return value == "" || strings.EqualFold(value, "true")
}

func addHTMLFromFS(ren multitemplate.Render, fs fs.FS, name string, files ...string) {
	tmpl := template.Must(template.New(files[0]).
		Funcs(templatefuncs.FuncsMap).
		ParseFS(fs, files...))
	ren.Add(name, tmpl)
}

func (s HTTPServer) Serve() error {
	log.Info().Uint16("port", s.cfg.HTTP.Port).Msg("Starting server.")
	return s.engine.Run(fmt.Sprintf(":%v", s.cfg.HTTP.Port))
}

// handlePostWebhook handles newreleases.io webhook post requests
func (s HTTPServer) handlePostWebhook(c *gin.Context) {
	// parse newreleases.io webhook
	var release Release
	if err := c.ShouldBindJSON(&release); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issueRef, err := ensureJiraIssue(s.jira, release, s.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go tryApplyChanges(s.jira, s.patcher, release, issueRef.IssueRef, s.cfg)

	// NOTE: always return OK, otherwise newreleases.io will retry
	c.Status(http.StatusOK)
}

func tryApplyChanges(j jira.Client, patcher patch.Patcher, release Release, issueRef jira.IssueRef, cfg *config.Config) {
	tmplCtx := patch.TemplateContext{
		Package:   release.Project,
		Version:   release.Version,
		JiraIssue: issueRef.Key,
	}

	pkg, ok := cfg.TryFindPackage(release.Project)
	if !ok {
		log.Info().Str("project", release.Project).Msg("No package patching config was found. Skipping patching.")
		createTemplatedComment(j, issueRef, cfg.Jira.Issue.Comments.NoConfig, tmplCtx)
		return
	}
	prs, err := patcher.CloneAndPublishAll(pkg.Repos, tmplCtx)
	if err != nil {
		log.Error().Err(err).Str("project", release.Project).Msg("Failed creating patches.")
		createTemplatedComment(j, issueRef, cfg.Jira.Issue.Comments.PRFailed, TemplateContextError{
			TemplateContext: tmplCtx,
			Error:           err.Error(),
		})
		return
	}
	if len(prs) == 0 {
		log.Warn().Str("project", release.Project).Msg("Found package config, but no repositories were patched.")
		createTemplatedComment(j, issueRef, cfg.Jira.Issue.Comments.NoPatches, tmplCtx)
		return
	}
	log.Info().
		Str("project", release.Project).
		Int("count", len(prs)).
		Msg("Successfully created PRs for update.")

	createTemplatedComment(j, issueRef, cfg.Jira.Issue.Comments.PRCreated, TemplateContextPullRequests{
		TemplateContext: tmplCtx,
		PullRequests:    prs,
	})
}

func createTemplatedComment(j jira.Client, issueRef jira.IssueRef, tmpl *config.Template, tmplCtx any) {
	comment, err := tmpl.Render(tmplCtx)
	if err != nil {
		log.Error().Err(err).Msg("Failed templating Jira issue comment.")
		return
	}

	if err := j.CreateIssueComment(issueRef, comment); err != nil {
		log.Error().Err(err).Msg("Failed creating Jira issue comment.")
	}
}

type TemplateContextError struct {
	patch.TemplateContext
	Error string
}

type TemplateContextPullRequests struct {
	patch.TemplateContext
	PullRequests []github.PullRequest
}

type newJiraIssue struct {
	jira.IssueRef
	Created bool
}

func ensureJiraIssue(j jira.Client, r Release, cfg *config.Config) (newJiraIssue, error) {
	existingIssues, err := j.FindIssuesForPackage(r.Project)
	if err != nil {
		return newJiraIssue{}, err
	}

	if len(existingIssues) == 0 {
		// no previous issues, create new jira issue
		i := r.JiraIssue(&cfg.Jira.Issue)

		if cfg.DryRun {
			log.Info().
				Str("issue", i.Key).
				Msg("Skipping creation of issue because Config.DryRun is enabled.")
			return newJiraIssue{
				IssueRef: i.IssueRef(),
				Created:  false,
			}, nil
		}
		issueRef, err := j.CreateIssue(i)
		if err != nil {
			return newJiraIssue{}, err
		}
		return newJiraIssue{
			IssueRef: issueRef,
			Created:  true,
		}, nil
	}

	// in case of duplicate issues, update the oldest (probably original) one, ignore rest as duplicates
	mostRecentIssue := existingIssues[0]
	var duplicateIssueKeys []string
	for _, issue := range existingIssues[1:] {
		duplicateIssueKeys = append(duplicateIssueKeys, issue.Key)
	}

	if len(duplicateIssueKeys) > 0 {
		log.Debug().
			Str("recent", mostRecentIssue.Key).
			Strs("duplicates", duplicateIssueKeys).
			Msg("Ignoring the duplicate issues in favor of recent issue.")
	}

	if cfg.DryRun {
		log.Info().
			Str("issue", mostRecentIssue.Key).
			Msg("Skipping update of issue because Config.DryRun is enabled.")
		return newJiraIssue{
			IssueRef: mostRecentIssue.IssueRef(),
			Created:  false,
		}, nil
	}
	issueRef := mostRecentIssue.IssueRef()
	if err := j.UpdateIssueSummary(issueRef, r.IssueSummary()); err != nil {
		return newJiraIssue{}, err
	}
	createTemplatedComment(j, issueRef, cfg.Jira.Issue.Comments.UpdatedIssue, patch.TemplateContext{
		Package:   r.Project,
		Version:   r.Version,
		JiraIssue: issueRef.Key,
	})
	return newJiraIssue{
		IssueRef: mostRecentIssue.IssueRef(),
		Created:  false,
	}, nil
}

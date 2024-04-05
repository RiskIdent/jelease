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
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/github"
	"github.com/RiskIdent/jelease/pkg/jira"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/RiskIdent/jelease/templates/pages"
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

func New(cfg *config.Config, j jira.Client, patcher patch.Patcher, staticFiles fs.FS) *HTTPServer {
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

	r.HTMLRender = &TemplRender{}

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", pages.Index())
	})

	r.GET("/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", pages.Config(cfg))
	})

	r.GET("/config/try-package", s.handleGetConfigTryPackage)
	r.POST("/config/try-package", s.handlePostConfigTryPackage)

	r.GET("/packages", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", pages.PackagesList(cfg))
	})

	r.GET("/packages/:package", func(c *gin.Context) {
		pkgName := c.Param("package")
		pkg, ok := s.cfg.TryFindPackage(pkgName)
		if !ok {
			c.HTML(http.StatusNotFound, "", pages.Error404(fmt.Sprintf("Package %q not found.", pkgName)))
			return
		}
		c.HTML(http.StatusOK, "", pages.PackagesItem(pages.PackageItemModel{
			Package: pkg,
		}))
	})

	r.GET("/packages/:package/create-pr", s.handleGetPRCreate)
	r.POST("/packages/:package/create-pr", s.handlePostPRCreate)

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/webhook") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			return
		}
		c.HTML(http.StatusNotFound, "", pages.Error404(""))
	})
	r.NoMethod(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/webhook") {
			var methodsAllowed []string
			for _, route := range r.Routes() {
				if route.Path == c.Request.URL.Path {
					methodsAllowed = append(methodsAllowed, route.Method)
				}
			}
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error":   "Method not allowed",
				"methods": methodsAllowed,
			})
			return
		}
		c.HTML(http.StatusNotFound, "", pages.Error405())
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
	tmplCtx := config.TemplateContext{
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

	if cfg.Jira.Issue.PRDeferredCreation {
		if cfg.HTTP.PublicURL == nil {
			log.Error().Msg("Cannot use deferred PR creation when no http.publicUrl is set. Falling back to creating PR automatically instead.")
		} else {
			u := createDeferredCreationURL(cfg.HTTP.PublicURL.URL(), pkg.Name, CreatePRRequest{
				Version:   release.Version,
				JiraIssue: issueRef.Key,
				PRCreate:  true,
			})

			createTemplatedComment(j, issueRef, cfg.Jira.Issue.Comments.PRDeferredCreation, TemplateContextURL{
				TemplateContext: tmplCtx,
				URL:             u,
			})
			return
		}
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
	createDynamicComment(j, issueRef, prs, release.Project, &cfg.Jira.Issue.Comments, tmplCtx)
}

func createDynamicComment(
	j jira.Client,
	issueRef jira.IssueRef,
	prs []github.PullRequest,
	pkgName string,
	commentTemplates *config.JiraIssueComments,
	tmplCtx config.TemplateContext,
) {
	if len(prs) == 0 {
		log.Warn().Str("project", pkgName).Msg("Found package config, but no repositories were patched.")
		createTemplatedComment(j, issueRef, commentTemplates.NoPatches, tmplCtx)
		return
	}
	log.Info().
		Str("project", pkgName).
		Int("count", len(prs)).
		Msg("Successfully created PRs for update.")

	createTemplatedComment(j, issueRef, commentTemplates.PRCreated, TemplateContextPullRequests{
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
	config.TemplateContext
	Error string
}

type TemplateContextPullRequests struct {
	config.TemplateContext
	PullRequests []github.PullRequest
}

type TemplateContextURL struct {
	config.TemplateContext
	URL *url.URL
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
	createTemplatedComment(j, issueRef, cfg.Jira.Issue.Comments.UpdatedIssue, config.TemplateContext{
		Package:   r.Project,
		Version:   r.Version,
		JiraIssue: issueRef.Key,
	})
	return newJiraIssue{
		IssueRef: mostRecentIssue.IssueRef(),
		Created:  false,
	}, nil
}

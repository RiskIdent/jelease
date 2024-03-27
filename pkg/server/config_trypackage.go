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
	"bytes"
	"errors"
	"net/http"
	"regexp"
	"text/template"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/patch"
	"github.com/RiskIdent/jelease/templates/pages"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ConfigTryPackageRequest is the query or form data pushed by the web.
type ConfigTryPackageRequest struct {
	Version       string `form:"version"`
	PackageConfig string `form:"config"`
}

func (s HTTPServer) bindConfigTryPackageRequest(c *gin.Context) (pages.ConfigTryPackageModel, bool) {
	var input ConfigTryPackageRequest
	err := c.ShouldBind(&input)
	model := pages.ConfigTryPackageModel{
		Config:        s.cfg,
		PackageConfig: input.PackageConfig,
		Version:       input.Version,
		IsPost:        c.Request.Method == http.MethodPost,
	}
	if yamlErr := yaml.Unmarshal([]byte(input.PackageConfig), &model.Package); yamlErr != nil {
		err = errors.Join(err, yamlErr)
	}
	if err != nil {
		model.Error = err
		c.HTML(http.StatusBadRequest, "", pages.ConfigTryPackage(model))
		return model, false
	}

	return model, true
}

// handleGetConfigTryPackage is the handler for:
//
//	GET /config/try-package
func (s HTTPServer) handleGetConfigTryPackage(c *gin.Context) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	enc.Encode(config.Package{
		Name: "example",
		Repos: []config.PackageRepo{
			{
				URL: "https://github.com/RiskIdent/jelease",
				Patches: []config.PackageRepoPatch{
					{
						Regex: &config.PatchRegex{
							File:    "go.mod",
							Match:   (*config.RegexPattern)(regexp.MustCompile(`^go [\d\.]+$`)),
							Replace: (*config.Template)(template.Must(template.New("").Parse("go {{ .Version }}"))),
						},
					},
				},
			},
		},
	})
	c.HTML(http.StatusOK, "", pages.ConfigTryPackage(pages.ConfigTryPackageModel{
		Config:        s.cfg,
		PackageConfig: buf.String(),
	}))
}

// handlePostConfigTryPackage is the handler for:
//
//	POST /config/try-package
func (s HTTPServer) handlePostConfigTryPackage(c *gin.Context) {
	model, ok := s.bindConfigTryPackageRequest(c)
	if !ok {
		return
	}

	cfgClone := *s.cfg
	cfgClone.DryRun = true
	patcherClone := s.patcher.CloneWithConfig(&cfgClone)

	tmplCtx := patch.TemplateContext{
		Package: model.Package.Name,
		Version: model.Version,
	}
	prs, err := patcherClone.CloneAndPublishAll(model.Package.Repos, tmplCtx)
	if err != nil {
		log.Error().Err(err).Str("project", model.Package.Name).Msg("Failed creating patches.")
	}

	model.PullRequests = prs
	model.Error = err
	c.HTML(http.StatusOK, "", pages.ConfigTryPackage(model))
}

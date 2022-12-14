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

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/jira"
)

// Release object unmarshaled from the newreleases.io webhook.
// Some fields omitted for simplicity, refer to the documentation at https://newreleases.io/webhooks
type Release struct {
	Provider string `json:"provider"`
	Project  string `json:"project"`
	Version  string `json:"version"`
}

// Generates a Textual summary for the release, intended to be used as the Jira issue summary
func (r Release) IssueSummary() string {
	return fmt.Sprintf("Update %v to version %v", r.Project, r.Version)
}

func (r Release) JiraIssue(cfg *config.JiraIssue) jira.Issue {
	return jira.Issue{
		Description:        cfg.Description,
		ProjectKey:         cfg.Project,
		TypeName:           cfg.Type,
		Labels:             cfg.Labels,
		Summary:            r.IssueSummary(),
		PackageName:        r.Project,
		PackageNameFieldID: cfg.ProjectNameCustomField,
	}
}

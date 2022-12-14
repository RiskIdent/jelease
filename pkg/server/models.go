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
	"github.com/andygrunwald/go-jira"
	"github.com/rs/zerolog/log"
	"github.com/trivago/tgo/tcontainer"
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

func (r Release) JiraIssue(cfg *config.Config) jira.Issue {
	labels := cfg.Jira.Issue.Labels
	var extraFields tcontainer.MarshalMap

	if cfg.Jira.Issue.ProjectNameCustomField == 0 {
		log.Trace().Msg("Create ticket with project name in labels.")
		labels = append(labels, r.Project)
	} else {
		log.Trace().
			Uint("customField", cfg.Jira.Issue.ProjectNameCustomField).
			Msg("Create ticket with project name in custom field.")
		customFieldName := fmt.Sprintf("customfield_%d", cfg.Jira.Issue.ProjectNameCustomField)
		extraFields = tcontainer.MarshalMap{
			customFieldName: r.Project,
		}
	}
	return jira.Issue{
		Fields: &jira.IssueFields{
			Description: cfg.Jira.Issue.Description,
			Project: jira.Project{
				Key: cfg.Jira.Issue.Project,
			},
			Type: jira.IssueType{
				Name: cfg.Jira.Issue.Type,
			},
			Labels:   labels,
			Summary:  r.IssueSummary(),
			Unknowns: extraFields,
		},
	}
}

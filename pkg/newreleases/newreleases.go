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

// An abstraction for interfacing with the newreleases.io api client
package newreleases

import (
	"context"
	"fmt"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/util"
	"gopkg.in/yaml.v3"
	"newreleases.io/newreleases"
)

// Abstraction over the newreleases go v1 api
type NewReleases struct {
	client        *newreleases.Client
	localProjects []config.NewReleasesProject
	defaults      config.NewReleasesDefaults
}

func FromCfg(cfg config.NewReleases) NewReleases {
	client := newreleases.NewClient(cfg.Auth.APIKey, nil)
	return NewReleases{
		client,
		cfg.Projects,
		cfg.Defaults,
	}
}

func (nr NewReleases) getProjects() ([]newreleases.Project, error) {
	var projects []newreleases.Project
	projects, lastPage, err := nr.client.Projects.List(context.Background(), newreleases.ProjectListOptions{Page: 1})
	if err != nil {
		return nil, err
	}
	// fetch additional pages if necessary
	for page := 2; page <= lastPage; page++ {
		options := newreleases.ProjectListOptions{Page: page}
		additionalProjects, _, err := nr.client.Projects.List(context.Background(), options)
		if err != nil {
			return nil, err
		}
		projects = append(projects, additionalProjects...)
	}
	return projects, nil
}

func mergeDefaults(projects []config.NewReleasesProject, nrDefaults config.NewReleasesDefaults) []config.NewReleasesProject {
	for i, project := range projects {
		if project.EmailNotification == "" {
			project.EmailNotification = nrDefaults.EmailNotification
			projects[i] = project
		}
	}
	return projects
}

func (nr NewReleases) Diff() (*ProjectDiff, error) {
	remoteProjects, err := nr.getProjects()
	if err != nil {
		return nil, err
	}
	remoteProjectCfgs := ProjectSliceToCfg(remoteProjects)
	localProjects := mapFromSlice(mergeDefaults(nr.localProjects, nr.defaults))
	diff := InitProjectDiff()

	for _, remoteProject := range remoteProjectCfgs {
		localProject, found := localProjects[remoteProject.Name]
		if !found {
			diff.MissingOnLocal[remoteProject.Name] = remoteProject
			continue
		} else if localProject.Equals(remoteProject) {
			diff.Identical[remoteProject.Name] = remoteProject
		} else {
			diff.Diverged[remoteProject.Name] = ProjectTuple{local: localProject, remote: remoteProject}
		}
		// clear all found projects from localProjects
		delete(localProjects, remoteProject.Name)
	}
	// all projects that have not been cleared from local must be missing on remote
	diff.MissingOnRemote = localProjects

	return &diff, nil
}

func mapFromSlice(slice []config.NewReleasesProject) map[string]config.NewReleasesProject {
	projectMap := make(map[string]config.NewReleasesProject, len(slice))
	for _, project := range slice {
		projectMap[project.Name] = project
	}
	return projectMap
}

func SliceFromMap(projectMap map[string]config.NewReleasesProject) []config.NewReleasesProject {
	projectSlice := make([]config.NewReleasesProject, len(projectMap))
	i := 0
	for _, project := range projectMap {
		projectSlice[i] = project
		i++
	}
	return projectSlice
}

func ExclusionToCfg(exclusion newreleases.Exclusion) config.NewReleasesExclusion {
	return config.NewReleasesExclusion{
		Value:   exclusion.Value,
		Inverse: exclusion.Inverse,
	}
}

func ExclusionFromCfg(exclusionCfg config.NewReleasesExclusion) newreleases.Exclusion {
	return newreleases.Exclusion{
		Value:   exclusionCfg.Value,
		Inverse: exclusionCfg.Inverse,
	}
}

func ExclusionSliceToCfg(exclusions []newreleases.Exclusion) []config.NewReleasesExclusion {
	exclusionConfigs := make([]config.NewReleasesExclusion, len(exclusions))
	for i, exclusion := range exclusions {
		exclusionConfigs[i] = ExclusionToCfg(exclusion)
	}
	return exclusionConfigs
}

func ExclusionSliceFromCfg(exclusions []config.NewReleasesExclusion) []newreleases.Exclusion {
	exclusionConfigs := make([]newreleases.Exclusion, len(exclusions))
	for i, exclusion := range exclusions {
		exclusionConfigs[i] = ExclusionFromCfg(exclusion)
	}
	return exclusionConfigs
}

func ProjectToCfg(project newreleases.Project) config.NewReleasesProject {
	return config.NewReleasesProject{
		Name:               project.Name,
		EmailNotification:  emailNotificationToString(project.EmailNotification),
		Provider:           project.Provider,
		Exclusions:         ExclusionSliceToCfg(project.Exclusions),
		ExcludePrereleases: project.ExcludePrereleases,
		ExcludeUpdated:     project.ExcludeUpdated,
	}
}

func ProjectSliceToCfg(projects []newreleases.Project) []config.NewReleasesProject {
	projectCfgs := make([]config.NewReleasesProject, len(projects))
	for i, project := range projects {
		projectCfgs[i] = ProjectToCfg(project)
	}
	return projectCfgs
}

// A diff over local and remote newreleases.io [ProjectCfg] configuration
// Contains hashmaps that have to be initialized. Create a new object with [NewProjectDiff]
type ProjectDiff struct {
	Identical       map[string]config.NewReleasesProject
	MissingOnLocal  map[string]config.NewReleasesProject
	MissingOnRemote map[string]config.NewReleasesProject
	Diverged        map[string]ProjectTuple
}

// Creates a ProjectDiff and initialized the contained maps
func InitProjectDiff() ProjectDiff {
	return ProjectDiff{
		Identical:       make(map[string]config.NewReleasesProject),
		MissingOnLocal:  make(map[string]config.NewReleasesProject),
		MissingOnRemote: make(map[string]config.NewReleasesProject),
		Diverged:        make(map[string]ProjectTuple),
	}
}

// Creates a formatted summary of the results of the diff
// Skips identical elements
func (diff ProjectDiff) Summary() string {
	var sb strings.Builder
	sb.WriteString("Missing on your newreleases.io account: (can be created with `jelease nr apply`)\n")
	for missingName, missingProject := range diff.MissingOnRemote {
		sb.WriteString(fmt.Sprintf("+ %s on %s\n", missingName, missingProject.Provider))
	}
	sb.WriteString("\n")
	sb.WriteString("Missing in local configuration: (can be imported with `jelease nr import`)\n")
	for missingName, missingProject := range diff.MissingOnLocal {
		sb.WriteString(fmt.Sprintf("? %s on %s\n", missingName, missingProject.Provider))
	}
	sb.WriteString("\n")
	sb.WriteString("These configurations have diverged: (requires manual fix)\n")
	for missingName := range diff.Diverged {
		sb.WriteString(fmt.Sprintf("! %s\n", missingName))
	}

	return sb.String()
}

func (diff ProjectDiff) DescribeDiverged() (string, error) {
	var sb strings.Builder
	for projectName, divergedTuple := range diff.Diverged {

		sb.WriteString(fmt.Sprintf("project %s\n", projectName))
		localYaml, err := yaml.Marshal(divergedTuple.local)
		if err != nil {
			return "", err
		}
		remoteYaml, err := yaml.Marshal(divergedTuple.remote)
		if err != nil {
			return "", err
		}
		sb.WriteString("local:\n")
		sb.WriteString(string(localYaml))
		sb.WriteString("remote:\n")
		sb.WriteString(string(remoteYaml))
	}

	return sb.String(), nil
}

type ProjectTuple struct {
	local, remote config.NewReleasesProject
}

type ApplyLocalConfigOptions struct {
	Destructive bool // indicates whether remote projects not present in local configuration should be removed
}

func (nr NewReleases) ApplyLocalConfig(options ApplyLocalConfigOptions) error {
	diff, err := nr.Diff()
	if err != nil {
		return err
	}

	for _, projectCfg := range diff.MissingOnRemote {
		emailNotification, err := emailNotificationFromString(projectCfg.EmailNotification)
		if err != nil {
			return err
		}
		projectOptions := newreleases.ProjectOptions{
			EmailNotification:  &emailNotification,
			WebhookIDs:         []string{},
			Exclusions:         ExclusionSliceFromCfg(projectCfg.Exclusions),
			ExcludePrereleases: &projectCfg.ExcludePrereleases,
			ExcludeUpdated:     &projectCfg.ExcludeUpdated,
			Note:               util.Ref(""),
			TagIDs:             []string{},
		}
		_, err = nr.client.Projects.Add(context.Background(), projectCfg.Provider, projectCfg.Name, &projectOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func emailNotificationFromString(notification string) (newreleases.EmailNotification, error) {
	var emailNotification newreleases.EmailNotification
	switch notification {
	case "none":
		emailNotification = newreleases.EmailNotificationNone
	case "instant":
		emailNotification = newreleases.EmailNotificationInstant
	case "hourly":
		emailNotification = newreleases.EmailNotificationHourly
	case "daily":
		emailNotification = newreleases.EmailNotificationDaily
	case "weekly":
		emailNotification = newreleases.EmailNotificationWeekly
	case "default":
		emailNotification = newreleases.EmailNotificationDefault
	default:
		return "", fmt.Errorf("failed to parse email configuration: %q is not valid value", notification)
	}
	return emailNotification, nil
}

func emailNotificationToString(notification newreleases.EmailNotification) string {
	var s string
	switch notification {
	case newreleases.EmailNotificationNone:
		s = "none"
	case newreleases.EmailNotificationInstant:
		s = "instant"
	case newreleases.EmailNotificationHourly:
		s = "hourly"
	case newreleases.EmailNotificationDaily:
		s = "daily"
	case newreleases.EmailNotificationWeekly:
		s = "weekly"
	case newreleases.EmailNotificationDefault:
		s = "default"
	default:
		s = "none"
	}
	return s
}

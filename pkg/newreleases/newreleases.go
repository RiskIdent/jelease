// An abstraction for interfacing with the newreleases.io api client
package newreleases

import (
	"context"
	"fmt"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"gopkg.in/yaml.v3"
	"newreleases.io/newreleases"
)

// Abstraction over the newreleases go v1 api
type NewReleases struct {
	client        *newreleases.Client
	localProjects []config.ProjectCfg
	defaults      config.NewReleasesDefaults
}

func FromCfg(cfg config.NewReleases) NewReleases {
	client := newreleases.NewClient(cfg.Auth.Key, nil)
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

func mergeDefaults(projects []config.ProjectCfg, nrDefaults config.NewReleasesDefaults) []config.ProjectCfg {
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

func mapFromSlice(slice []config.ProjectCfg) map[string]config.ProjectCfg {
	projectMap := make(map[string]config.ProjectCfg, len(slice))
	for _, project := range slice {
		projectMap[project.Name] = project
	}
	return projectMap
}

func SliceFromMap(projectMap map[string]config.ProjectCfg) []config.ProjectCfg {
	projectSlice := make([]config.ProjectCfg, len(projectMap))
	i := 0
	for _, project := range projectMap {
		projectSlice[i] = project
		i++
	}
	return projectSlice
}

func ExclusionToCfg(exclusion newreleases.Exclusion) config.ExclusionCfg {
	return config.ExclusionCfg{
		Value:   exclusion.Value,
		Inverse: exclusion.Inverse,
	}
}

func ExclusionFromCfg(exclusionCfg config.ExclusionCfg) newreleases.Exclusion {
	return newreleases.Exclusion{
		Value:   exclusionCfg.Value,
		Inverse: exclusionCfg.Inverse,
	}
}

func ExclusionSliceToCfg(exclusions []newreleases.Exclusion) []config.ExclusionCfg {
	exclusionConfigs := make([]config.ExclusionCfg, len(exclusions))
	for i, exclusion := range exclusions {
		exclusionConfigs[i] = ExclusionToCfg(exclusion)
	}
	return exclusionConfigs
}

func ExclusionSliceFromCfg(exclusions []config.ExclusionCfg) []newreleases.Exclusion {
	exclusionConfigs := make([]newreleases.Exclusion, len(exclusions))
	for i, exclusion := range exclusions {
		exclusionConfigs[i] = ExclusionFromCfg(exclusion)
	}
	return exclusionConfigs
}

func ProjectToCfg(project newreleases.Project) config.ProjectCfg {
	return config.ProjectCfg{
		Name:               project.Name,
		EmailNotification:  string(project.EmailNotification),
		Provider:           project.Provider,
		WebhookIDs:         project.WebhookIDs,
		Exclusions:         ExclusionSliceToCfg(project.Exclusions),
		ExcludePrereleases: project.ExcludePrereleases,
		ExcludeUpdated:     project.ExcludeUpdated,
		Note:               project.Note,
	}
}

func ProjectSliceToCfg(projects []newreleases.Project) []config.ProjectCfg {
	projectCfgs := make([]config.ProjectCfg, len(projects))
	for i, project := range projects {
		projectCfgs[i] = ProjectToCfg(project)
	}
	return projectCfgs
}

// A diff over local and remote newreleases.io [ProjectCfg] configuration
// Contains hashmaps that have to be initialized. Create a new object with [NewProjectDiff]
type ProjectDiff struct {
	Identical       map[string]config.ProjectCfg
	MissingOnLocal  map[string]config.ProjectCfg
	MissingOnRemote map[string]config.ProjectCfg
	Diverged        map[string]ProjectTuple
}

// Creates a ProjectDiff and initialized the contained maps
func InitProjectDiff() ProjectDiff {
	return ProjectDiff{
		Identical:       make(map[string]config.ProjectCfg),
		MissingOnLocal:  make(map[string]config.ProjectCfg),
		MissingOnRemote: make(map[string]config.ProjectCfg),
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
	local, remote config.ProjectCfg
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
		projectOptions := newreleases.ProjectOptions{
			EmailNotification:  (*newreleases.EmailNotification)(&projectCfg.EmailNotification),
			WebhookIDs:         projectCfg.WebhookIDs,
			Exclusions:         ExclusionSliceFromCfg(projectCfg.Exclusions),
			ExcludePrereleases: &projectCfg.ExcludePrereleases,
			ExcludeUpdated:     &projectCfg.ExcludeUpdated,
			Note:               &projectCfg.Note,
			TagIDs:             []string{},
		}
		_, err := nr.client.Projects.Add(context.Background(), projectCfg.Provider, projectCfg.Name, &projectOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

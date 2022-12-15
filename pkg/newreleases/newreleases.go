// An abstraction for interfacing with the newreleases.io api client
package newreleases

import (
	"context"
	"fmt"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"newreleases.io/newreleases"
)

// Abstraction over the with the v1 newreleases go api
type NewReleases struct {
	client        *newreleases.Client
	localProjects []config.ProjectCfg
}

func FromCfg(cfg config.NewReleases) NewReleases {
	client := newreleases.NewClient(cfg.Auth.Key, nil)
	return NewReleases{
		client,
		cfg.Projects,
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

func (nr NewReleases) Diff() (*ProjectDiff, error) {
	remoteProjects, err := nr.getProjects()
	if err != nil {
		return nil, err
	}
	remoteProjectCfgs := ProjectCfgSliceFromProjectSlice(remoteProjects)
	localProjects := mapFromSlice(nr.localProjects)
	diff := NewProjectDiff()

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

func ExclusionCfgFromExclusion(exclusion newreleases.Exclusion) config.ExclusionCfg {
	return config.ExclusionCfg{
		Value:   exclusion.Value,
		Inverse: exclusion.Inverse,
	}
}

func ExclusionCfgSliceFromExclusionSlice(exclusions []newreleases.Exclusion) []config.ExclusionCfg {
	exclusionConfigs := make([]config.ExclusionCfg, len(exclusions))
	for i, exclusion := range exclusions {
		exclusionConfigs[i] = ExclusionCfgFromExclusion(exclusion)
	}
	return exclusionConfigs
}

func ProjectCfgFromProject(project newreleases.Project) config.ProjectCfg {
	return config.ProjectCfg{
		Name:                   project.Name,
		Provider:               project.Provider,
		URL:                    project.URL,
		EmailNotification:      string(project.EmailNotification),
		SlackIDs:               project.SlackIDs,
		TelegramChatIDs:        project.TelegramChatIDs,
		DiscordIDs:             project.DiscordIDs,
		HangoutsChatWebhookIDs: project.HangoutsChatWebhookIDs,
		MSTeamsWebhookIDs:      project.MSTeamsWebhookIDs,
		MattermostWebhookIDs:   project.MattermostWebhookIDs,
		RocketchatWebhookIDs:   project.RocketchatWebhookIDs,
		MatrixRoomIDs:          project.MatrixRoomIDs,
		WebhookIDs:             project.WebhookIDs,
		Exclusions:             ExclusionCfgSliceFromExclusionSlice(project.Exclusions),
		ExcludePrereleases:     project.ExcludePrereleases,
		ExcludeUpdated:         project.ExcludeUpdated,
		Note:                   project.Note,
	}
}

func ProjectCfgSliceFromProjectSlice(projects []newreleases.Project) []config.ProjectCfg {
	projectCfgs := make([]config.ProjectCfg, len(projects))
	for i, exclusion := range projects {
		projectCfgs[i] = ProjectCfgFromProject(exclusion)
	}
	return projectCfgs
}

// type Project = newreleases.Project

// A diff over local and remote newreleases.io [Project] configurations
// Contains hashmaps that have to be initialized. Create a new object with [NewProjectDiff]
type ProjectDiff struct {
	Identical       map[string]config.ProjectCfg
	MissingOnLocal  map[string]config.ProjectCfg
	MissingOnRemote map[string]config.ProjectCfg
	Diverged        map[string]ProjectTuple
}

// Creates a ProjectDiff and initialized the contained maps
func NewProjectDiff() ProjectDiff {
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
	for missingName, _ := range diff.Diverged {
		sb.WriteString(fmt.Sprintf("! %s\n", missingName))
	}

	return sb.String()
}

func (diff ProjectDiff) DescribeDiverged() string {
	var sb strings.Builder
	for projectName, divergedTuple := range diff.Diverged {
		sb.WriteString(fmt.Sprintf("project %s\n", projectName))
		sb.WriteString(fmt.Sprintf("local:\n%v\nremote:\n%v\n---\n", divergedTuple.local, divergedTuple.remote))
	}

	return sb.String()
}

type ProjectTuple struct {
	local, remote config.ProjectCfg
}

func (nr NewReleases) applyLocalConfig() {

}

func (nr NewReleases) ImportProjects() (string, error) {
	diff, err := nr.Diff()
	if err != nil {
		return "", err
	}
	for _, project := range diff.MissingOnLocal {

		fmt.Println(project)
		// yaml.Marshal()
		// jsonBytes, err := json.Marshal(project)
		// if err != nil {
		// 	return "", err
		// }
		// jsonString := string(jsonBytes)
		// // we don't care about the Id, remove
		// withoutId, err := sjson.Delete(jsonString, "id")
		// if err != nil {
		// 	return "", err
		// }
		// yamlBytes, err := yaml.JSONToYAML([]byte(withoutId))
		// if err != nil {
		// 	return "", err
		// }
		// yamlString := string(yamlBytes)
		// fmt.Println(yamlString)
	}

	// TODO: fix return type
	return "", nil
}

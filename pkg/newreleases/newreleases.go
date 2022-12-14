// An abstraction for interfacing with the newreleases.io api client
package newreleases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"newreleases.io/newreleases"
)

// Abstraction over the with the v1 newreleases go api
type NewReleases struct {
	client        *newreleases.Client
	localProjects []Project
}

func FromCfg(apiKey string, projects []Project) NewReleases {
	client := newreleases.NewClient(apiKey, nil)
	return NewReleases{
		client,
		projects,
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
	localProjects := mapFromSlice(nr.localProjects)

	diff := NewProjectDiff()

	for _, remoteProject := range remoteProjects {
		localProject, found := localProjects[remoteProject.Name]
		if !found {
			diff.MissingOnLocal[remoteProject.Name] = remoteProject
			continue
		}
		// diff on json because it's difficult in go
		// TODO: find way to ignore ID, which won't be present locally
		localJson, err := json.Marshal(localProject)
		if err != nil {
			return nil, err
		}
		remoteJson, err := json.Marshal(remoteProject)
		if err != nil {
			return nil, err
		}
		localString := string(localJson)
		remoteString := string(remoteJson)

		if localString != remoteString {
			diff.Diverged[remoteProject.Name] = ProjectTuple{local: localProject, remote: remoteProject}
			continue
		}
		diff.Identical[remoteProject.Name] = remoteProject
	}

	// TODO: add missing remotely
	return &diff, nil
}

func mapFromSlice(slice []Project) map[string]Project {
	projectMap := make(map[string]Project, len(slice))
	for _, project := range slice {
		projectMap[project.Name] = project
	}
	return projectMap
}

// func equal(left Project, right Project) {

// }

type Project = newreleases.Project

// A diff over local and remote newreleases.io [Project] configurations
// Contains hashmaps that have to be initialized. Create a new object with [NewProjectDiff]
type ProjectDiff struct {
	Identical       map[string]Project
	MissingOnLocal  map[string]Project
	MissingOnRemote map[string]Project
	Diverged        map[string]ProjectTuple
}

// Creates a ProjectDiff and initialized the contained maps
func NewProjectDiff() ProjectDiff {
	return ProjectDiff{
		Identical:       make(map[string]newreleases.Project),
		MissingOnLocal:  make(map[string]newreleases.Project),
		MissingOnRemote: make(map[string]newreleases.Project),
		Diverged:        make(map[string]ProjectTuple),
	}
}

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

type ProjectTuple struct {
	local, remote Project
}

func (nr NewReleases) applyLocalConfig() {

}

func (nr NewReleases) importProjects() {

}

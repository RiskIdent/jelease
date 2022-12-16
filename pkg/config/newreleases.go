package config

import "golang.org/x/exp/slices"

type NewReleases struct {
	Auth     NewReleasesAuth
	Projects []ProjectCfg
	Defaults NewReleasesDefaults
	Webhook  string
}

type NewReleasesAuth struct {
	Key string
}

type NewReleasesDefaults struct {
	EmailNotification string `yaml:"emailNotifications"`
}

// type EmailNotificationCfg string

// Omits some field of [newreleases.io/newreleases/Project] that we don't want to store
// namely, this omits the ID field and the tagIDs field
type ProjectCfg struct {
	Name               string
	Provider           string
	EmailNotification  string         `json:",omitempty" yaml:",omitempty"`
	WebhookIDs         []string       `json:",omitempty" yaml:",omitempty"`
	Exclusions         []ExclusionCfg `json:"excludeVersionRegexp,omitempty" yaml:"excludeVersionRegexp,omitempty" mapstructure:"excludeVersionRegexp"`
	ExcludePrereleases bool           `json:"excludePrereleases,omitempty" yaml:"excludePrereleases,omitempty" mapstructure:"excludePrereleases"`
	ExcludeUpdated     bool           `json:"excludeUpdated,omitempty" yaml:"excludeUpdated,omitempty" mapstructure:"excludeUpdated"`
	Note               string         `json:",omitempty" yaml:",omitempty"`
	// TagIDs                 []string             `json:"tags,omitempty"`
}

type ExclusionCfg struct {
	Value   string `json:"value"`
	Inverse bool   `json:"inverse"`
}

func (exclusion ExclusionCfg) Equals(other ExclusionCfg) bool {
	return exclusion.Value == other.Value && exclusion.Inverse == other.Inverse
}

func ExclusionCfgSliceEquals(left []ExclusionCfg, right []ExclusionCfg) bool {
	if len(left) != len(right) {
		return false
	}
	for i, value := range left {
		if right[i] != value {
			return false
		}
	}
	return true
}

// Thanks to Go, I now know how a compiler feels
func (project ProjectCfg) Equals(other ProjectCfg) bool {
	if project.Name != other.Name {
		return false
	}
	if project.Provider != other.Provider {
		return false
	}
	if project.EmailNotification != other.EmailNotification {
		return false
	}
	if !slices.Equal(project.WebhookIDs, other.WebhookIDs) {
		return false
	}
	if !ExclusionCfgSliceEquals(project.Exclusions, other.Exclusions) {
		return false
	}
	if project.ExcludePrereleases != other.ExcludePrereleases {
		return false
	}
	if project.ExcludeUpdated != other.ExcludeUpdated {
		return false
	}
	if project.Note != other.Note {
		return false
	}
	return true
}

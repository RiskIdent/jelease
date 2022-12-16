package config

import "golang.org/x/exp/slices"

type NewReleases struct {
	Auth              NewReleasesAuth      `yaml:"auth"`
	Projects          []ProjectCfg         `yaml:"projects"`
	EmailNotification EmailNotificationCfg `yaml:"emailNotifications"`
}

type NewReleasesAuth struct {
	Key string
}

type EmailNotificationCfg string

// Omits some field of [newreleases.io/newreleases/Project] that we don't want to store
// namely, this omits the ID field and the tagIDs field
type ProjectCfg struct {
	Name                   string
	Provider               string
	URL                    string         `json:"url" yaml:"url" mapstructure:"url"`
	EmailNotification      string         `json:",omitempty" yaml:",omitempty"`
	SlackIDs               []string       `json:",omitempty" yaml:",omitempty"`
	TelegramChatIDs        []string       `json:",omitempty" yaml:",omitempty"`
	DiscordIDs             []string       `json:",omitempty" yaml:",omitempty"`
	HangoutsChatWebhookIDs []string       `json:",omitempty" yaml:",omitempty"`
	MSTeamsWebhookIDs      []string       `json:",omitempty" yaml:",omitempty"`
	MattermostWebhookIDs   []string       `json:",omitempty" yaml:",omitempty"`
	RocketchatWebhookIDs   []string       `json:",omitempty" yaml:",omitempty"`
	MatrixRoomIDs          []string       `json:",omitempty" yaml:",omitempty"`
	WebhookIDs             []string       `json:",omitempty" yaml:",omitempty"`
	Exclusions             []ExclusionCfg `json:"excludeVersionRegexp,omitempty" yaml:"excludeVersionRegexp,omitempty" mapstructure:"excludeVersionRegexp"`
	ExcludePrereleases     bool           `json:",omitempty" yaml:",omitempty"`
	ExcludeUpdated         bool           `json:",omitempty" yaml:",omitempty"`
	Note                   string         `json:",omitempty" yaml:",omitempty"`
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
	if project.URL != other.URL {
		return false
	}
	if project.EmailNotification != other.EmailNotification {
		return false
	}
	if !slices.Equal(project.SlackIDs, other.SlackIDs) {
		return false
	}
	if !slices.Equal(project.TelegramChatIDs, other.TelegramChatIDs) {
		return false
	}
	if !slices.Equal(project.DiscordIDs, other.DiscordIDs) {
		return false
	}
	if !slices.Equal(project.HangoutsChatWebhookIDs, other.HangoutsChatWebhookIDs) {
		return false
	}
	if !slices.Equal(project.MSTeamsWebhookIDs, other.MSTeamsWebhookIDs) {
		return false
	}
	if !slices.Equal(project.MattermostWebhookIDs, other.MattermostWebhookIDs) {
		return false
	}
	if !slices.Equal(project.RocketchatWebhookIDs, other.RocketchatWebhookIDs) {
		return false
	}
	if !slices.Equal(project.MatrixRoomIDs, other.MatrixRoomIDs) {
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

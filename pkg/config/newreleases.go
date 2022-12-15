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
	Name                   string         `json:"name" yaml:"name"`
	Provider               string         `json:"provider" yaml:"provider"`
	URL                    string         `json:"url" yaml:"url"`
	EmailNotification      string         `json:"email_notification,omitempty" yaml:"email_notification,omitempty"`
	SlackIDs               []string       `json:"slack_channels,omitempty" yaml:"slack_channels,omitempty"`
	TelegramChatIDs        []string       `json:"telegram_chats,omitempty" yaml:"telegram_chats,omitempty"`
	DiscordIDs             []string       `json:"discord_channels,omitempty" yaml:"discord_channels,omitempty"`
	HangoutsChatWebhookIDs []string       `json:"hangouts_chat_webhooks,omitempty" yaml:"hangouts_chat_webhooks,omitempty"`
	MSTeamsWebhookIDs      []string       `json:"microsoft_teams_webhooks,omitempty" yaml:"microsoft_teams_webhooks,omitempty"`
	MattermostWebhookIDs   []string       `json:"mattermost_webhooks,omitempty" yaml:"mattermost_webhooks,omitempty"`
	RocketchatWebhookIDs   []string       `json:"rocketchat_webhooks,omitempty" yaml:"rocketchat_webhooks,omitempty"`
	MatrixRoomIDs          []string       `json:"matrix_rooms,omitempty" yaml:"matrix_rooms,omitempty"`
	WebhookIDs             []string       `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
	Exclusions             []ExclusionCfg `json:"exclude_version_regexp,omitempty" yaml:"exclude_version_regexp,omitempty"`
	ExcludePrereleases     bool           `json:"exclude_prereleases,omitempty" yaml:"exclude_prereleases,omitempty"`
	ExcludeUpdated         bool           `json:"exclude_updated,omitempty" yaml:"exclude_updated,omitempty"`
	Note                   string         `json:"note,omitempty" yaml:"note,omitempty"`
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
	if project.ExcludeUpdated != other.ExcludePrereleases {
		return false
	}
	if project.Note != other.Note {
		return false
	}
	return true
}

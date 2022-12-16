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
	EmailNotification string `json:"emailNotification" yaml:"emailNotification" mapstructure:"emailNotification"`
}

// Similar to [newreleases.io/newreleases/Project] but omits some fields that we don't want to store or compare
type ProjectCfg struct {
	Name               string
	Provider           string
	EmailNotification  string         `json:"emailNotification,omitempty" yaml:"emailNotification,omitempty" mapstructure:"emailNotification"`
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

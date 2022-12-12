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

package cmd

import "testing"

func TestNewJiraIssueSearchQuery(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		project     string
		customField uint
		want        string
	}{
		{
			name:        "no custom field",
			status:      "Grooming",
			project:     "platform/jelease",
			customField: 0,
			want:        `status = "Grooming" and labels = "platform/jelease"`,
		},
		{
			name:        "with custom field",
			status:      "Grooming",
			project:     "platform/jelease",
			customField: 12500,
			want:        `status = "Grooming" and (labels = "platform/jelease" or cf[12500] ~ "platform/jelease")`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := newJiraIssueSearchQuery(tc.status, tc.project, tc.customField)
			if tc.want != got {
				t.Errorf("Wrong query.\nwant: `%s`\ngot:  `%s`", tc.want, got)
			}
		})
	}
}

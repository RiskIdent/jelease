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

package git

import "testing"

func TestAddCredentialsToURL(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		user   string
		pass   string
		want   string
	}{
		{
			name:   "with cred",
			remote: "https://github.com/RiskIdent/jelease.git",
			user:   "my-user",
			pass:   "my-pass",
			want:   "https://my-user:my-pass@github.com/RiskIdent/jelease.git",
		},
		{
			name:   "without cred",
			remote: "https://github.com/RiskIdent/jelease.git",
			want:   "https://github.com/RiskIdent/jelease.git",
		},
		{
			name:   "only username",
			remote: "https://github.com/RiskIdent/jelease.git",
			user:   "my-user",
			want:   "https://my-user@github.com/RiskIdent/jelease.git",
		},
		{
			name:   "only password",
			remote: "https://github.com/RiskIdent/jelease.git",
			pass:   "my-pass",
			want:   "https://:my-pass@github.com/RiskIdent/jelease.git",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := addCredentialsToRemote(tc.remote, Credentials{
				Username: tc.user,
				Password: tc.pass,
			})
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != got {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

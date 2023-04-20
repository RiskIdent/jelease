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

package main

import (
	"embed"
	"fmt"

	"github.com/RiskIdent/jelease/cmd"
	"github.com/RiskIdent/jelease/pkg/config"
	"gopkg.in/yaml.v3"
)

//go:embed jelease.yaml
var defaultConfigYAML []byte

//go:embed templates
var templatesFS embed.FS

func main() {
	var defaultConfig config.Config
	if err := yaml.Unmarshal(defaultConfigYAML, &defaultConfig); err != nil {
		panic(fmt.Errorf("Parse embedded config: %w", err))
	}
	cmd.Execute(defaultConfig, templatesFS)
}

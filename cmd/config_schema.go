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

import (
	"encoding/json"
	"fmt"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/spf13/cobra"
)

var configSchemaFlags = struct {
	indented bool
}{
	indented: true,
}

// configSchemaCmd represents the config command
var configSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Prints the JSON schema for the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := config.Schema()
		data, err := marshalJSON(s, configSchemaFlags.indented)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Intentionally overrides the config loading from root.go
		return nil
	},
}

func marshalJSON(v any, indented bool) ([]byte, error) {
	if indented {
		return json.MarshalIndent(v, "", "  ")
	}
	return json.Marshal(v)
}

func init() {
	configCmd.AddCommand(configSchemaCmd)

	configSchemaCmd.Flags().BoolVarP(&configSchemaFlags.indented, "indent", "i", configSchemaFlags.indented, "Print indented output")
}

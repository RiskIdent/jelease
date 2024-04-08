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

package patches

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/rs/zerolog/log"
)

func ApplyHelmDepUpdatePatch(repoDir string, tmplCtx config.TemplateContext, patch config.PatchHelmDepUpdate) error {
	if patch.Chart == nil {
		return fmt.Errorf("missing required field 'chart'")
	}

	chart, err := patch.Chart.ExecuteString(tmplCtx)
	if err != nil {
		return fmt.Errorf("execute chart dir template: %w", err)
	}

	log.Info().Str("chart", chart).Msg("Executing `helm dependency update`")
	cmd := exec.Command("helm", "dependency", "update")
	cmd.Dir = filepath.Join(repoDir, chart)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w; command output:\n%s", err, out)
	}

	return nil
}

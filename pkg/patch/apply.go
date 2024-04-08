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

package patch

import (
	"errors"
	"fmt"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/patch/filestore"
	"github.com/RiskIdent/jelease/pkg/patch/patches"
)

// ApplyMany applies a series of patches in sequence using [Apply].
func ApplyMany(repoDir string, patches []config.PackageRepoPatch, tmplCtx config.TemplateContext) error {
	for _, p := range patches {
		if err := Apply(repoDir, p, tmplCtx); err != nil {
			return err
		}
	}
	return nil
}

// Apply applies a single patch to the repository.
func Apply(repoDir string, patch config.PackageRepoPatch, tmplCtx config.TemplateContext) error {
	fstore := filestore.NewCached(repoDir)
	defer fstore.Close()
	switch {
	case patch.Regex != nil:
		if err := patches.ApplyRegexPatch(fstore, tmplCtx, *patch.Regex); err != nil {
			return fmt.Errorf("regex patch: %w", err)
		}
	case patch.YAML != nil:
		if err := patches.ApplyYAMLPatch(fstore, tmplCtx, *patch.YAML); err != nil {
			return fmt.Errorf("yaml patch: %w", err)
		}
	case patch.HelmDepUpdate != nil:
		// Flush the store as we need the up-to-date changes on disk
		if err := fstore.Flush(); err != nil {
			return err
		}
		if err := patches.ApplyHelmDepUpdatePatch(repoDir, tmplCtx, *patch.HelmDepUpdate); err != nil {
			return fmt.Errorf("exec patch: %w", err)
		}
	default:
		return errors.New("missing patch type config")
	}

	return fstore.Close()
}

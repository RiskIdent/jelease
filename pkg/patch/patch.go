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
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/rs/zerolog/log"
)

type TemplateContext struct {
	Package   string
	Version   string
	JiraIssue string
}

type TemplateContextRegex struct {
	TemplateContext
	Groups []string
}

// ApplyMany applies a series of patches in sequence using [Apply].
func ApplyMany(repoDir string, patches []config.PackageRepoPatch, tmplCtx TemplateContext) error {
	for _, p := range patches {
		if err := Apply(repoDir, p, tmplCtx); err != nil {
			return err
		}
	}
	return nil
}

// Apply applies a single patch to the repository.
func Apply(repoDir string, patch config.PackageRepoPatch, tmplCtx TemplateContext) error {
	fstore := NewCachedFileStore(repoDir)
	defer fstore.Close()
	switch {
	case patch.Regex != nil:
		if err := applyRegexPatch(fstore, tmplCtx, *patch.Regex); err != nil {
			return fmt.Errorf("regex patch: %w", err)
		}
	case patch.YQ != nil:
		return errors.New("yq patch: not implemented")
	default:
		return errors.New("missing patch type config")
	}

	return fstore.Close()
}

func applyRegexPatch(fstore FileStore, tmplCtx TemplateContext, patch config.PatchRegex) error {
	log.Debug().Str("file", patch.File).Stringer("match", patch.Match).Msg("Patching regex.")

	content, err := fstore.ReadFile(patch.File)
	if err != nil {
		return err
	}
	regex := patch.Match.Regexp()
	lines := bytes.Split(content, []byte("\n"))

	for i, line := range lines {
		groupIndices := regex.FindSubmatchIndex(line)
		if groupIndices == nil {
			// No match
			continue
		}

		fullMatchStart := groupIndices[0]
		fullMatchEnd := groupIndices[1]

		everythingBefore := line[:fullMatchStart]
		everythingAfter := line[fullMatchEnd:]

		var buf bytes.Buffer
		if err := patch.Replace.Template().Execute(&buf, TemplateContextRegex{
			TemplateContext: tmplCtx,
			Groups:          regexSubmatchIndicesToStrings(line, groupIndices),
		}); err != nil {
			return fmt.Errorf("line %d: execute replace template: %w", i+1, err)
		}
		lines[i] = slices.Concat(everythingBefore, buf.Bytes(), everythingAfter)
		newContent := bytes.Join(lines, []byte("\n"))

		return fstore.WriteFile(patch.File, newContent)
	}

	return fmt.Errorf("regex did not match any line: %s", patch.Match)
}

func regexSubmatchIndicesToStrings(line []byte, indices []int) []string {
	strs := make([]string, 0, len(indices)/2)
	for i := 0; i < len(indices); i += 2 {
		start := indices[i]
		end := indices[i+1]
		strs = append(strs, string(line[start:end]))
	}
	return strs
}

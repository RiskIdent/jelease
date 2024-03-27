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
	"os"
	"path/filepath"
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
			return fmt.Errorf("file %q: %w", p.File, err)
		}
	}
	return nil
}

// Apply applies a single patch to the repository.
func Apply(repoDir string, patch config.PackageRepoPatch, tmplCtx TemplateContext) error {
	log.Debug().Str("file", patch.File).Msg("Patching file.")

	// TODO: Check that the patch path doesn't go outside the repo dir.
	// For example, reject stuff like "../../../somefile.txt"
	path := filepath.Join(repoDir, patch.File)
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	switch {
	case patch.Regex != nil:
		content, err = applyRegexPatch(content, tmplCtx, *patch.Regex)
	default:
		return errors.New("missing patch type config")
	}
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, stat.Mode())
}

func applyRegexPatch(content []byte, tmplCtx TemplateContext, patch config.PatchRegex) ([]byte, error) {
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
			return nil, fmt.Errorf("line %d: execute replace template: %w", i+1, err)
		}
		lines[i] = slices.Concat(everythingBefore, buf.Bytes(), everythingAfter)
		return bytes.Join(lines, []byte("\n")), nil
	}

	return nil, fmt.Errorf("regex did not match any line: %s", patch.Match)
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

func writeLines(path string, lines [][]byte) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, stat.Mode())
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range lines {
		for len(line) > 0 {
			n, err := file.Write(line)
			if err != nil {
				return err
			}
			if n == 0 {
				return errors.New("wrote 0 bytes, stopping infinite loop")
			}
			line = line[n:]
		}
		file.Write([]byte("\n"))
	}
	return nil
}

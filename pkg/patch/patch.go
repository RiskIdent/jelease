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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/rs/zerolog/log"
)

type TemplateContext struct {
	Package string
	Version string
}

type TemplateContextRegex struct {
	TemplateContext
	Groups []string
}

func Apply(repoDir string, patch config.PackageRepoPatch, tmplCtx TemplateContext) error {
	// TODO: Check that the patch path doesn't go outside the repo dir.
	// For example, reject stuff like "../../../somefile.txt"
	path := filepath.Join(repoDir, patch.File)

	lines, err := readLines(path)
	if err != nil {
		return fmt.Errorf("read file for patch: %w", err)
	}

	if err := patchLines(patch, tmplCtx, lines); err != nil {
		return fmt.Errorf("patch lines: %w", err)
	}

	if err := writeLines(path, lines); err != nil {
		return fmt.Errorf("write patch: %w", err)
	}

	log.Info().Str("file", patch.File).Msg("Patched file.")
	return nil
}

func patchLines(patch config.PackageRepoPatch, tmplCtx TemplateContext, lines [][]byte) error {
	for i, line := range lines {
		newLine, err := patchSingleLine(patch, tmplCtx, line)
		if err != nil {
			return err
		}
		if newLine == nil { // No match
			continue
		}
		if bytes.Equal(line, newLine) {
			return errors.New("found matching line, but already up-to-date")
		}
		lines[i] = newLine
		return nil // Stop after first match
	}
	return errors.New("no match in file")
}

func patchSingleLine(patch config.PackageRepoPatch, tmplCtx TemplateContext, line []byte) ([]byte, error) {
	switch {
	case patch.Regex != nil:
		return patchSingleLineRegex(*patch.Regex, tmplCtx, line)
	default:
		return nil, errors.New("missing patch type config")
	}
}

func patchSingleLineRegex(patch config.PatchRegex, tmplCtx TemplateContext, line []byte) ([]byte, error) {
	regex := patch.Match.Regexp()
	groupIndices := regex.FindSubmatchIndex(line)
	if groupIndices == nil {
		// No match
		return nil, nil
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
		return nil, fmt.Errorf("execute replace template: %w", err)
	}

	return util.Concat(everythingBefore, buf.Bytes(), everythingAfter), nil
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

func readLines(path string) ([][]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readLinesFromReader(file)
}

func readLinesFromReader(r io.Reader) ([][]byte, error) {
	var lines [][]byte
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

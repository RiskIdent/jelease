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
	"bytes"
	"fmt"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/patch/filestore"
	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type TemplateContextRegex struct {
	config.TemplateContext
	Groups []string
}

func ApplyYAMLPatch(fstore filestore.FileStore, tmplCtx config.TemplateContext, patch config.PatchYAML) error {
	log.Debug().Str("file", patch.File).Stringer("yamlpath", patch.YAMLPath).Msg("Patching YAML.")

	if patch.File == "" {
		return fmt.Errorf("missing required field 'file'")
	}
	if patch.YAMLPath == nil {
		return fmt.Errorf("missing required field 'yamlPath'")
	}
	if patch.Replace == nil {
		return fmt.Errorf("missing required field 'replace'")
	}

	content, err := fstore.ReadFile(patch.File)
	if err != nil {
		return err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return err
	}

	matches, err := patch.YAMLPath.YAMLPath.Find(&node)
	if err != nil {
		return fmt.Errorf("yamlpath %q: eval: %w", patch.YAMLPath, err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("yamlpath %q: no matches found", patch.YAMLPath)
	}

	if patch.MaxMatches > 0 && len(matches) > patch.MaxMatches {
		return fmt.Errorf("yamlpath %q: matched too many times: %d, max = %d", patch.YAMLPath, len(matches), patch.MaxMatches)
	}

	for _, match := range matches {
		if match.ShortTag() != "!!str" {
			return fmt.Errorf("yamlpath %q: line %d: only supports matching strings, but instead matched %q", patch.YAMLPath, match.Line, match.ShortTag())
		}
		var buf bytes.Buffer
		if err := patch.Replace.Template().Execute(&buf, tmplCtx); err != nil {
			return fmt.Errorf("yamlpath %q: line %d: execute replace template: %w", patch.YAMLPath, match.Line, err)
		}
		setYAMLNodeRecursive(match, buf.String())
	}

	newContent, err := yamlEncode(&node, patch.Indent)
	if err != nil {
		return err
	}

	fixedContent, err := util.PatchKeepWhitespace(content, newContent)
	switch {
	case err != nil:
		log.Debug().
			Err(err).
			Str("file", patch.File).
			Stringer("yamlpath", patch.YAMLPath).
			Msg("Failed to perform whitespace preserving patch. Skipping that step.")
	case bytes.Equal(fixedContent, newContent):
		log.Debug().
			Str("file", patch.File).
			Stringer("yamlpath", patch.YAMLPath).
			Msg("No changes applied via whitespace preserving patch.")
	default:
		log.Debug().
			Str("file", patch.File).
			Stringer("yamlpath", patch.YAMLPath).
			Msg("Applied whitespace preserving patch.")
		newContent = fixedContent
	}

	return fstore.WriteFile(patch.File, newContent)
}

func setYAMLNodeRecursive(node *yaml.Node, value string) {
	if node.Alias != nil {
		setYAMLNodeRecursive(node.Alias, value)
		return
	}
	node.SetString(value)
}

func yamlEncode(obj any, indent int) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	if indent > 0 {
		enc.SetIndent(indent)
	} else {
		enc.SetIndent(2)
	}

	if err := enc.Encode(obj); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

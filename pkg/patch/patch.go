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
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
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
	case patch.YAML != nil:
		if err := applyYAMLPatch(fstore, tmplCtx, *patch.YAML); err != nil {
			return fmt.Errorf("yaml patch: %w", err)
		}
	case patch.Exec != nil:
		// Flush the store as we need the up-to-date changes on disk
		if err := fstore.Flush(); err != nil {
			return err
		}
		if err := applyExecPatch(repoDir, tmplCtx, *patch.Exec); err != nil {
			return fmt.Errorf("exec patch: %w", err)
		}
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

func applyYAMLPatch(fstore FileStore, tmplCtx TemplateContext, patch config.PatchYAML) error {
	log.Debug().Str("file", patch.File).Stringer("yamlpath", patch.YAMLPath).Msg("Patching YAML.")

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

func applyExecPatch(repoDir string, tmplCtx TemplateContext, patch config.PatchExec) error {
	cmd, err := templateExecCommand(repoDir, tmplCtx, patch)
	if err != nil {
		return fmt.Errorf("create command: %w", err)
	}
	log.Debug().Stringer("cmd", cmd).Msg("Patching exec.")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w; command output:\n%s", err, out)
	}

	return nil
}

func templateExecCommand(repoDir string, tmplCtx TemplateContext, patch config.PatchExec) (*exec.Cmd, error) {
	if len(patch.Cmd) == 0 {
		return nil, fmt.Errorf("missing command")
	}
	cmdStrings := make([]string, len(patch.Cmd))
	var buf bytes.Buffer
	for i, tmpl := range patch.Cmd {
		if tmpl == nil {
			return nil, fmt.Errorf("arg %d: arg must not be null", i)
		}
		buf.Reset()
		if err := tmpl.Template().Execute(&buf, tmplCtx); err != nil {
			return nil, fmt.Errorf("arg %d: execute template: %w", i, err)
		}
		cmdStrings[i] = buf.String()
	}
	cmd := exec.Command(cmdStrings[0], cmdStrings[1:]...)
	cmd.Dir = repoDir
	if patch.Dir != "" {
		cmd.Dir = filepath.Join(repoDir, patch.Dir)
	}
	envStrings := os.Environ()
	envMap := make(map[string]string, len(envStrings))
	for _, envStr := range envStrings {
		key, value, _ := strings.Cut(envStr, "=")
		envMap[key] = value
	}
	for key, tmpl := range patch.Env {
		buf.Reset()
		if err := tmpl.Template().Execute(&buf, tmplCtx); err != nil {
			return nil, fmt.Errorf("env %q: execute template: %w", key, err)
		}
		envMap[key] = buf.String()
	}
	envStrings = make([]string, 0, len(envMap))
	for key, value := range envMap {
		envStrings = append(envStrings, fmt.Sprintf("%s=%s", key, value))
	}
	cmd.Env = envStrings
	return cmd, nil
}

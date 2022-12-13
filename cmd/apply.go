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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:  "apply <package> <version>",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgName := args[0]
		version := args[1]
		pkg, ok := tryFindPackageConfig(pkgName)
		if !ok {
			return fmt.Errorf("no such package found in config: %s", pkgName)
		}
		log.Info().Str("package", pkgName).Msg("Found package config")

		for _, patch := range pkg.Patches {
			log.Info().Str("file", patch.File).Msg("Patching file")
			if err := applyPatch(patch, pkg.Name, version); err != nil {
				return err
			}
		}

		log.Info().Str("package", pkgName).Msg("Done applying patches")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

func tryFindPackageConfig(pkgName string) (config.Package, bool) {
	for _, pkg := range cfg.Packages {
		if pkg.Name == pkgName {
			return pkg, true
		}
	}
	return config.Package{}, false
}

func applyPatch(patch config.PackagePatch, pkgName, version string) error {
	g := git.Cmd{Credentials: git.Credentials{}}
	repo, err := prepareRepo(g, patch.Repo)
	if err != nil {
		return err
	}
	// TODO: uncomment
	//defer repo.Close()

	if err := applyPatchToRepo(repo, patch, version); err != nil {
		return err
	}

	return commitAndPushChanges(g, repo, pkgName, version)
}

func applyPatchToRepo(repo git.Repo, patch config.PackagePatch, version string) error {
	path := filepath.Join(repo.Directory(), patch.File)

	lines, err := readLines(path)
	if err != nil {
		return fmt.Errorf("read file for patch: %w", err)
	}

	if err := patchLines(patch, version, lines); err != nil {
		return fmt.Errorf("patch lines: %w", err)
	}

	if err := writeLines(path, lines); err != nil {
		return fmt.Errorf("write patch: %w", err)
	}

	log.Info().Str("file", patch.File).Msg("Patched file.")
	return nil
}

func patchLines(patch config.PackagePatch, version string, lines [][]byte) error {
	for i, line := range lines {
		newLine, err := patchSingleLine(patch, version, line)
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

func patchSingleLine(patch config.PackagePatch, version string, line []byte) ([]byte, error) {
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
	if err := patch.Replace.Template().Execute(&buf, struct {
		Groups  []string
		Version string
	}{
		Groups:  regexSubmatchIndicesToStrings(line, groupIndices),
		Version: version,
	}); err != nil {
		return nil, fmt.Errorf("execute replace template: %w", err)
	}

	return concat(everythingBefore, buf.Bytes(), everythingAfter), nil
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

func concat[S ~[]E, E any](slices ...S) S {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	result := make(S, totalLen)
	var offset int
	for _, s := range slices {
		copy(result[offset:], s)
		offset += len(s)
	}
	return result
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

func prepareRepo(g git.Git, repoURL string) (git.Repo, error) {
	dir, err := os.MkdirTemp("tmp", "jelease-repo-*")
	if err != nil {
		return nil, err
	}
	repo, err := g.Clone(dir, repoURL)
	if err != nil {
		return nil, err
	}
	log.Info().Str("branch", repo.CurrentBranch()).Str("dir", repo.Directory()).Msg("Cloned repo.")
	if err := repo.CheckoutNewBranch("jelease/is/awesome"); err != nil {
		return nil, err
	}
	log.Info().Str("branch", repo.CurrentBranch()).Str("mainBranch", repo.MainBranch()).Msg("Checked out new branch.")
	return repo, nil
}

func commitAndPushChanges(g git.Git, repo git.Repo, pkgName, version string) error {
	if err := repo.StageChanges(); err != nil {
		return err
	}
	log.Info().Msg("Staged changes.")

	if err := repo.CreateCommit(fmt.Sprintf("Updated %v to %v", pkgName, version)); err != nil {
		return err
	}
	log.Info().Msg("Commit created.")

	if err := repo.PushChanges(); err != nil {
		return err
	}
	log.Info().Msg("Pushed changes.")
	return nil
}
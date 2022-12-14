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
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/fatih/color"
	"github.com/google/go-github/v48/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
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

		if len(pkg.Repos) == 0 {
			log.Warn().Str("package", pkgName).Msg("No repos configured for package.")
			return nil
		}

		for _, pkgRepo := range pkg.Repos {
			log.Info().Str("repo", pkgRepo.URL).Msg("Patching repo")
			if err := applyRepoPatches(pkgRepo, pkg.Name, version); err != nil {
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

func applyRepoPatches(pkgRepo config.PackageRepo, pkgName, version string) error {
	if len(pkgRepo.Patches) == 0 {
		log.Warn().Str("package", pkgName).Str("repo", pkgRepo.URL).Msg("No patches configured for repository.")
	}

	// Check this early so we don't fail right on the finish line
	repoRef, err := getGitHubRepoRef(pkgRepo.URL)
	if err != nil {
		return err
	}

	g := git.Cmd{Credentials: git.Credentials{}}
	repo, err := prepareRepo(g, pkgRepo.URL, pkgName, version)
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Warn().Err(err).Str("dir", repo.Directory()).
				Msg("Failed to clean up cloned temporary repo directory.")
		} else {
			log.Debug().Str("dir", repo.Directory()).
				Msg("Cleaned up cloned temporary repo directory.")
		}
	}()

	for _, patch := range pkgRepo.Patches {
		if err := applyPatchToRepo(repo, patch, version); err != nil {
			return err
		}
	}

	if err := commitAndPushChanges(g, repo, pkgName, version); err != nil {
		return err
	}

	return createPR(repo, repoRef, pkgName, version)
}

func applyPatchToRepo(repo git.Repo, patch config.PackageRepoPatch, version string) error {
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

func patchLines(patch config.PackageRepoPatch, version string, lines [][]byte) error {
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

func patchSingleLine(patch config.PackageRepoPatch, version string, line []byte) ([]byte, error) {
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

func prepareRepo(g git.Git, repoURL, pkgName, version string) (git.Repo, error) {
	dir, err := createRepoTempDirectory()
	if err != nil {
		return nil, err
	}
	repo, err := g.Clone(dir, repoURL)
	if err != nil {
		return nil, err
	}
	log.Info().Str("branch", repo.CurrentBranch()).Str("dir", repo.Directory()).Msg("Cloned repo.")
	branchName, err := cfg.GitHub.PR.Branch.Render(struct {
		Package string
		Version string
	}{
		Package: pkgName,
		Version: version,
	})
	if err != nil {
		return nil, fmt.Errorf("template branch name: %w", err)
	}
	if err := repo.CheckoutNewBranch(branchName); err != nil {
		return nil, err
	}
	log.Info().Str("branch", repo.CurrentBranch()).Str("mainBranch", repo.MainBranch()).Msg("Checked out new branch.")
	return repo, nil
}

func createRepoTempDirectory() (string, error) {
	parentDir := filepath.Join(deref(cfg.GitHub.TempDir, os.TempDir()), "jelease-cloned-repos")
	if err := os.MkdirAll(parentDir, 0700); err != nil {
		return "", err
	}
	return os.MkdirTemp(parentDir, "jelease-repo-*")
}

func commitAndPushChanges(g git.Git, repo git.Repo, pkgName, version string) error {
	logDiff(repo)

	if err := repo.StageChanges(); err != nil {
		return err
	}
	log.Debug().Msg("Staged changes.")

	commit, err := repo.CreateCommit(fmt.Sprintf("Updated %v to %v", pkgName, version))
	if err != nil {
		return err
	}
	log.Debug().
		Str("hash", commit.AbbrHash).
		Str("subject", commit.Subject).
		Msg("Created commit.")

	if cfg.DryRun {
		log.Info().Msg("Dry run: skipping pushing changes.")
		return nil
	}

	if err := repo.PushChanges(); err != nil {
		return err
	}
	log.Info().Msg("Pushed changes to remote repository.")
	return nil
}

func logDiff(repo git.Repo) {
	if log.Logger.GetLevel() > zerolog.DebugLevel {
		return
	}
	diff, err := repo.DiffChanges()
	if err != nil {
		log.Warn().Err(err).Msg("Failed diffing changes. Trying to continue anyways.")
		return
	}
	if !color.NoColor && cfg.Log.Format == config.LogFormatPretty {
		diff = colorizeDiff(diff)
	}
	log.Debug().Msgf("Diff:\n%s", diff)
}

var (
	colorDiffTrippleDash   = color.New(color.FgHiRed, color.Italic)
	colorDiffRemove        = color.New(color.FgRed)
	colorDiffTripplePlus   = color.New(color.FgHiGreen, color.Italic)
	colorDiffAdd           = color.New(color.FgGreen)
	colorDiffDoubleAt      = color.New(color.FgMagenta, color.Italic)
	colorDiffOtherNonSpace = color.New(color.FgHiBlack, color.Italic)
)

func colorizeDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "---"):
			lines[i] = colorDiffTrippleDash.Sprint(line)
		case strings.HasPrefix(line, "-"):
			lines[i] = colorDiffRemove.Sprint(line)
		case strings.HasPrefix(line, "+++"):
			lines[i] = colorDiffTripplePlus.Sprint(line)
		case strings.HasPrefix(line, "+"):
			lines[i] = colorDiffAdd.Sprint(line)
		case strings.HasPrefix(line, "@@"):
			lines[i] = colorDiffDoubleAt.Sprint(line)
		case !strings.HasPrefix(line, " "):
			lines[i] = colorDiffOtherNonSpace.Sprint(line)
		}
	}
	return strings.Join(lines, "\n")
}

func createPR(repo git.Repo, repoRef GitHubRepoRef, pkgName, version string) error {
	gh, err := newGitHubClient()
	if err != nil {
		return fmt.Errorf("new GitHub client: %w", err)
	}

	tmplData := struct {
		Package string
		Version string
	}{
		Package: pkgName,
		Version: version,
	}
	title, err := cfg.GitHub.PR.Title.Render(tmplData)
	if err != nil {
		return fmt.Errorf("template PR title: %w", err)
	}
	description, err := cfg.GitHub.PR.Description.Render(tmplData)
	if err != nil {
		return fmt.Errorf("template PR description: %w", err)
	}

	if cfg.DryRun {
		log.Info().Msg("Dry run: skipping creating GitHub pull request.")
		return nil
	}

	pr, _, err := gh.PullRequests.Create(context.TODO(), repoRef.Owner, repoRef.Repo, &github.NewPullRequest{
		Title:               &title,
		Body:                &description,
		Head:                ref(repo.CurrentBranch()),
		Base:                ref(repo.MainBranch()),
		MaintainerCanModify: ref(true),
	})
	if err != nil {
		return fmt.Errorf("create GitHub PR: %w", err)
	}
	log.Info().
		Int("pr", deref(pr.Number, -1)).
		Str("url", deref(pr.HTMLURL, "")).
		Msg("GitHub PR created.")
	return nil
}

func ref[T any](v T) *T {
	return &v
}

func deref[T any](v *T, or T) T {
	if v == nil {
		return or
	}
	return *v
}

func newGitHubClient() (*github.Client, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GitHub.Auth.Token})
	tc := oauth2.NewClient(context.TODO(), ts)
	if cfg.GitHub.URL != nil {
		return github.NewEnterpriseClient(*cfg.GitHub.URL, "", tc)
	}
	return github.NewClient(tc), nil
}

type GitHubRepoRef struct {
	Owner string
	Repo  string
}

func getGitHubRepoRef(remote string) (GitHubRepoRef, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return GitHubRepoRef{}, err
	}
	u.User = nil
	path := u.Path
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) < 2 {
		return GitHubRepoRef{}, fmt.Errorf("expected https://host/OWNER/REPO in URL, got: %s", u.String())
	}
	return GitHubRepoRef{
		Owner: segments[0],
		Repo:  strings.TrimSuffix(segments[1], ".git"),
	}, nil
}

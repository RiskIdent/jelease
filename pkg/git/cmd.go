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

package git

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/RiskIdent/jelease/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Cmd implements [Git] using the command-line version of Git.
type Cmd struct {
	Credentials Credentials
	Committer   Committer
}

var _ Git = Cmd{}

func addCredentialsToRemote(remote string, cred Credentials) (string, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return "", fmt.Errorf("parse remote URL: %w", err)
	}
	switch {
	case cred.Password == "" && cred.Username == "":
		return remote, nil
	case cred.Password == "":
		u.User = url.User(cred.Username)
	default:
		u.User = url.UserPassword(cred.Username, cred.Password)
	}
	return u.String(), nil
}

func (g Cmd) Clone(targetDir, remote string) (Repo, error) {
	log.Debug().Str("dir", targetDir).Str("remote", remote).Msg("Cloning into dir")
	remoteWithCred, err := addCredentialsToRemote(remote, g.Credentials)
	if err != nil {
		return nil, err
	}
	_, err = runGitCmd("clone", "--single-branch", "--depth", "1", "--", remoteWithCred, targetDir)
	if err != nil {
		return nil, fmt.Errorf("clone repo: %w", err)
	}
	branchOutput, err := runAsCommitterInDir(g.Committer, targetDir, "branch", "--show-current")
	if err != nil {
		return nil, fmt.Errorf("check current branch: %w", err)
	}
	branchName := strings.TrimSpace(string(branchOutput))
	return &CmdRepo{
		Committer:     g.Committer,
		directory:     targetDir,
		currentBranch: branchName,
		mainBranch:    branchName,
	}, nil
}

type CmdRepo struct {
	Committer     Committer
	directory     string
	currentBranch string
	mainBranch    string
}

var _ Repo = &CmdRepo{}

func (r *CmdRepo) run(args ...string) ([]byte, error) {
	return runAsCommitterInDir(r.Committer, r.directory, args...)
}

func (r *CmdRepo) Directory() string {
	return r.directory
}

func (r *CmdRepo) CurrentBranch() string {
	return r.currentBranch
}

func (r *CmdRepo) MainBranch() string {
	return r.mainBranch
}

func (r *CmdRepo) CheckoutNewBranch(branchName string) error {
	_, err := r.run("checkout", "-b", branchName)
	if err != nil {
		return fmt.Errorf("checkout branch: %s: %w", branchName, err)
	}
	r.currentBranch = branchName
	return nil
}

func (r *CmdRepo) DiffChanges() (string, error) {
	output, err := r.run("diff")
	if err != nil {
		return "", fmt.Errorf("diff changes: %w", err)
	}
	return string(output), nil
}

func (r *CmdRepo) DiffStaged() (string, error) {
	output, err := r.run("diff", "--staged")
	if err != nil {
		return "", fmt.Errorf("diff changes: %w", err)
	}
	return string(output), nil
}

func (r *CmdRepo) StageChanges() error {
	_, err := r.run("add", "--all")
	if err != nil {
		return fmt.Errorf("stage all changes: %w", err)
	}
	return nil
}

func (r *CmdRepo) CreateCommit(message string) (Commit, error) {
	diff, err := r.DiffStaged()
	if err != nil {
		log.Warn().Err(err).Msg("Failed diffing changes. Trying to continue anyways.")
		diff = "# Failed to diff. See console output"
	}
	if _, err := r.run("commit", "-m", message, "--no-gpg-sign"); err != nil {
		return Commit{}, fmt.Errorf("commit changes: %w", err)
	}
	output, err := r.run("show", "--no-notes", "--no-patch", "--format=%H%n%h%n%P%n%p%n%s")
	if err != nil {
		return Commit{}, fmt.Errorf("get commit details: %w", err)
	}
	lines := bytes.Split(output, []byte("\n"))
	if len(lines) < 5 {
		return Commit{}, fmt.Errorf("get commit details: expected 5 lines, got %d", len(lines))
	}
	return Commit{
		Hash:           string(lines[0]),
		AbbrHash:       string(lines[1]),
		ParentHash:     string(lines[2]),
		ParentAbbrHash: string(lines[3]),
		Subject:        string(lines[4]),
		Diff:           diff,
	}, nil
}

func (r *CmdRepo) PushChanges() error {
	_, err := r.run("push", "--set-upstream", "origin", r.currentBranch)
	if err != nil {
		return fmt.Errorf("push changes: %w", err)
	}
	return nil
}

func (r *CmdRepo) Close() error {
	return os.RemoveAll(r.directory)
}

func runAsCommitterInDir(committer Committer, targetDir string, args ...string) ([]byte, error) {
	extraArgs := []string{"-C", targetDir}
	if committer.Name != "" {
		extraArgs = append(extraArgs, "-c", "user.name="+committer.Name)
	}
	if committer.Email != "" {
		extraArgs = append(extraArgs, "-c", "user.email="+committer.Email)
	}
	return runGitCmd(util.Concat(extraArgs, args)...)
}

func runGitCmd(args ...string) ([]byte, error) {
	if log.Logger.GetLevel() <= zerolog.DebugLevel {
		log.Trace().Strs("args", censorArgs(args)).Msg("Executing Git command.")
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("%w, output: %s", err, exitErr.Stderr)
		}
		return nil, err
	}
	return output, nil
}

func censorArgs(args []string) []string {
	censored := make([]string, len(args))
	copy(censored, args)
	for i, arg := range censored {
		switch {
		case strings.HasPrefix(arg, "user.name="):
			censored[i] = "user.name=..."
		case strings.HasPrefix(arg, "user.email="):
			censored[i] = "user.email=..."
		case strings.HasPrefix(arg, "https://"),
			strings.HasPrefix(arg, "http://"):
			u, err := url.Parse(arg)
			if err != nil {
				continue
			}
			if u.User != nil {
				u.User = url.UserPassword("...", "...")
			}
			censored[i] = u.String()
		}
	}
	return censored
}

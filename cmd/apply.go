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
	"os"
	"path/filepath"

	"github.com/RiskIdent/jelease/pkg/git"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use: "apply",
	RunE: func(cmd *cobra.Command, args []string) error {
		g := git.Cmd{Credentials: git.Credentials{}}
		dir, err := os.MkdirTemp("tmp", "jelease-repo-*")
		if err != nil {
			return err
		}

		//repo, err := g.Clone(dir, "https://github.com/RiskIdent/jelease.git")
		repo, err := g.Clone(dir, "tmp/upstream-test")
		if err != nil {
			return err
		}
		//defer repo.Close()
		log.Info().Str("branch", repo.CurrentBranch()).Str("dir", repo.Directory()).Msg("Cloned repo.")
		if err := repo.CheckoutNewBranch("jelease/is/awesome"); err != nil {
			return err
		}
		log.Info().Str("branch", repo.CurrentBranch()).Str("mainBranch", repo.MainBranch()).Msg("Checked out new branch.")

		if err := createTestFile(filepath.Join(repo.Directory(), "testfile.txt")); err != nil {
			return err
		}
		log.Info().Msg("Created test file")

		if err := repo.StageChanges(); err != nil {
			return err
		}
		log.Info().Msg("Staged changes.")

		if err := repo.CreateCommit("Some fancy commit message"); err != nil {
			return err
		}
		log.Info().Msg("Commit created.")

		if err := repo.PushChanges(); err != nil {
			return err
		}
		log.Info().Msg("Pushed changes.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

func createTestFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("Lorem ipsum")
	return err
}

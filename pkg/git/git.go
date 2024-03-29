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
	"io"
	"os"
	"path/filepath"
)

type Git interface {
	Clone(targetDir, remote string) (Repo, error)
}

type Repo interface {
	io.Closer
	Directory() string
	CurrentBranch() string
	MainBranch() string

	CheckoutNewBranch(branchName string) error
	DiffChanges() (string, error)
	DiffStaged() (string, error)
	StageChanges() error
	CreateCommit(message string) (Commit, error)
	PushChanges() error
}

type Credentials struct {
	Username string
	Password string
}

type Committer struct {
	Name  string // maps to Git config user.name
	Email string // maps to Git config user.email
}

type Commit struct {
	Hash           string
	AbbrHash       string
	ParentHash     string
	ParentAbbrHash string
	Subject        string
	Diff           string
}

func (c Commit) String() string {
	return c.Hash
}

func CloneTemp(g Git, tmpDirPattern, remote string) (Repo, error) {
	parentDir, filePattern := filepath.Split(tmpDirPattern)
	if err := os.MkdirAll(parentDir, 0700); err != nil {
		return nil, err
	}
	tmpDir, err := os.MkdirTemp(parentDir, filePattern)
	if err != nil {
		return nil, err
	}
	return g.Clone(tmpDir, remote)
}

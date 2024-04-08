// SPDX-FileCopyrightText: 2024 Risk.Ident GmbH <contact@riskident.com>
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

package filestore

import "os"

func NewTestFileStore(files map[string]string) *TestFileStore {
	return &TestFileStore{
		files: files,
	}
}

// TestFileStore is a [FileStore] used during Go unit tests.
// It never touches the OS' underlying file system.
type TestFileStore struct {
	files map[string]string
}

// ensure it implements the interface
var _ FileStore = &TestFileStore{}

func (s *TestFileStore) ReadFile(path string) ([]byte, error) {
	content, ok := s.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}

func (s *TestFileStore) WriteFile(path string, content []byte) error {
	s.files[path] = string(content)
	return nil
}

func (s *TestFileStore) Close() error {
	return nil
}

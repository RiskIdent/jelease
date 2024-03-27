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

package patch

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FileStore interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte) error
	Close() error
}

func NewCachedFileStore(dir string) *CachedFileStore {
	return &CachedFileStore{
		Dir:   dir,
		files: map[string]*File{},
	}
}

type CachedFileStore struct {
	Dir   string
	files map[string]*File
}

// ensure it implements the interface
var _ FileStore = &CachedFileStore{}

type File struct {
	Content []byte
	Mode    fs.FileMode
}

func (s *CachedFileStore) ReadFile(path string) ([]byte, error) {
	path = filepath.Clean(path)
	if file, ok := s.files[path]; ok {
		return file.Content, nil
	}
	// TODO: Check that the patch path doesn't go outside the repo dir.
	// For example, reject stuff like "../../../somefile.txt"
	absPath := filepath.Join(s.Dir, path)
	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	s.files[path] = &File{content, stat.Mode()}
	return content, nil
}

func (s *CachedFileStore) WriteFile(path string, content []byte) error {
	path = filepath.Clean(path)
	if _, ok := s.files[path]; !ok {
		// Load it
		_, err := s.ReadFile(path)
		if err != nil {
			return err
		}
	}
	s.files[path].Content = content
	return nil
}

func (s *CachedFileStore) Close() error {
	for path, file := range s.files {
		if err := os.WriteFile(filepath.Join(s.Dir, path), file.Content, file.Mode); err != nil {
			return err
		}
	}
	s.files = map[string]*File{}
	return nil
}

func NewTestFileStore(files map[string]string) *TestFileStore {
	return &TestFileStore{
		files: files,
	}
}

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

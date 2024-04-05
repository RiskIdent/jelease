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

import (
	"io/fs"
	"os"
	"path/filepath"
)

func NewCached(dir string) *Cached {
	return &Cached{
		Dir:   dir,
		files: map[string]*File{},
	}
}

// Cached is a [FileStore] that uses the OS' file system,
// but adds in-memory caching in between.
// The files are never written to disk until the [Cached.Flush]
// or [FileStore.Close] method is called.
type Cached struct {
	Dir   string
	files map[string]*File
}

// ensure it implements the interface
var _ FileStore = &Cached{}

type File struct {
	Content []byte
	Mode    fs.FileMode
}

func (s *Cached) ReadFile(path string) ([]byte, error) {
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

func (s *Cached) WriteFile(path string, content []byte) error {
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

func (s *Cached) Flush() error {
	for path, file := range s.files {
		if err := os.WriteFile(filepath.Join(s.Dir, path), file.Content, file.Mode); err != nil {
			return err
		}
	}
	s.files = map[string]*File{}
	return nil
}

func (s *Cached) Close() error {
	return s.Flush()
}

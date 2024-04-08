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

package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// patchKeepWhitespace applies changes while ignoring whitespace changes.
// Used as the YAML libraries trims away the whitespace, so this function
// translates the changes from patches while ignoring whitespace trimming.
//
// Based on: [https://github.com/mikefarah/yq/issues/515#issuecomment-1574420861]
//
// Effectively doing:
//
//	cat "$modified_file" | diff -Bw "$original" - | patch "$original" -
//
// Note that this approach is not 100% and does not work when the edits
// are surrounded by whitespace.
// Instead, this only preserves the whitespace unaffected by the edits.
func PatchKeepWhitespace(original, modified []byte) ([]byte, error) {
	originalPath, err := writeTemp("patch-original-file-*", original)
	if err != nil {
		return nil, fmt.Errorf("write to temp file: %w", err)
	}
	defer os.Remove(originalPath)

	// Cannot use the long flag variants, as we want to support busybox's diff as well.
	//  -B  Ignore changes whose lines are all blank
	//  -w  Ignore all whitespace
	diff := exec.Command("diff", "-Bw", originalPath, "-")
	diff.Stdin = bytes.NewReader(modified)
	diffStdout, err := diff.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout from 'diff': %w", err)
	}

	patch := exec.Command("patch", originalPath, "--input=-", "--output=-")
	patch.Stdin = diffStdout
	var patchStdoutBuf bytes.Buffer
	patch.Stdout = &patchStdoutBuf

	var wg sync.WaitGroup

	wg.Add(1)
	var diffErr error
	go func() {
		defer wg.Done()
		if err := execRunStderr(diff); err != nil {
			diffErr = fmt.Errorf("diff: %w", err)
		}
	}()

	wg.Add(1)
	var patchErr error
	go func() {
		defer wg.Done()
		if err := execRunStderr(patch); err != nil {
			patchErr = fmt.Errorf("patch: %w", err)
		}
	}()

	wg.Wait()

	if diffErr != nil {
		var exitErr *exec.ExitError
		if errors.As(diffErr, &exitErr) && exitErr.ProcessState.ExitCode() == 1 {
			// exit code 1 is OK. From 'diff --help':
			// 	"Exit status is 0 if inputs are the same, 1 if different, 2 if trouble."
		} else {
			return nil, diffErr
		}
	}
	if patchErr != nil {
		return nil, patchErr
	}

	return patchStdoutBuf.Bytes(), nil
}

func writeTemp(namePattern string, content []byte) (string, error) {
	tempDir := filepath.Join(os.TempDir(), "jelease")
	if err := os.MkdirAll(tempDir, 0o644); err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	f, err := os.CreateTemp(tempDir, namePattern)
	if err != nil {
		return "", err
	}
	defer f.Close()
	path := f.Name()
	if _, err := f.Write(content); err != nil {
		os.Remove(path)
		return path, err
	}
	return path, nil
}

func execRunStderr(cmd *exec.Cmd) error {
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("open stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	stderr, err := io.ReadAll(stderrReader)
	if err != nil {
		return fmt.Errorf("read stderr pipe: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w; stderr:\n%s", err, stderr)
	}
	return nil
}

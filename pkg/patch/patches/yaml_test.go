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

package patches

import (
	"testing"

	"github.com/RiskIdent/jelease/pkg/config"
	"github.com/RiskIdent/jelease/pkg/patch/filestore"
)

func TestApplYAMLPatch(t *testing.T) {
	fstore := filestore.NewTestFileStore(map[string]string{
		"file.txt": "{foo: bar}\n",
	})
	patch := config.PatchYAML{
		File:     "file.txt",
		YAMLPath: newYAMLPath(t, `.foo`),
		Replace:  newTemplate(t, `{{ .Package }}@{{ .Version }}`),
	}

	tmplCtx := config.TemplateContext{
		Package: "my-pkg",
		Version: "v1.2.3",
	}

	err := ApplyYAMLPatch(fstore, tmplCtx, patch)
	if err != nil {
		t.Fatal(err)
	}

	want := "{foo: my-pkg@v1.2.3}\n"
	gotBytes, err := fstore.ReadFile("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)

	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

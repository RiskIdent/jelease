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
	"strings"

	"github.com/fatih/color"
)

// Color settings for Git diffs.
//
// Example diff:
//
//	diff --git a/cmd/apply.go b/cmd/apply.go
//	index b4c2d30..a5f6e5f 100644
//	--- a/cmd/apply.go
//	+++ b/cmd/apply.go
//	@@ -339,42 +338,12 @@ func logDiff(repo git.Repo) {
//	 		log.Warn().Err(err).Msg("Failed diffing changes. Trying to continue anyways.")
//	 		return
//	 	}
//	-	if !color.NoColor && cfg.Log.Format == config.LogFormatPretty {
//	-		diff = colorizeDiff(diff)
//	+	if cfg.Log.Format == config.LogFormatPretty {
//	+		diff = git.ColorizeDiff(diff)
var (
	// ColorDiffTrippleDash is the "file name before change", e.g:
	//	--- a/cmd/apply.go
	ColorDiffTrippleDash = color.New(color.FgHiRed, color.Italic)
	// ColorDiffTripplePlus is the "file name after change", e.g:
	//	+++ b/cmd/apply.go
	ColorDiffTripplePlus = color.New(color.FgHiGreen, color.Italic)
	// ColorDiffRemove is the removed lines, e.g:
	//	-	if !color.NoColor && cfg.Log.Format == config.LogFormatPretty {
	//	-		diff = colorizeDiff(diff)
	ColorDiffRemove = color.New(color.FgRed)
	// ColorDiffAdd is the added lines, e.g:
	//	+	if cfg.Log.Format == config.LogFormatPretty {
	//	+		diff = git.ColorizeDiff(diff)
	ColorDiffAdd = color.New(color.FgGreen)
	// ColorDiffDoubleAt is the double at-symbol (@) specifying section ranges, e.g:
	//	@@ -339,42 +338,12 @@ func logDiff(repo git.Repo) {
	ColorDiffDoubleAt = color.New(color.FgMagenta, color.Italic)
	// ColorDiffOtherNonSpace is the other Git info lines, i.e all other lines
	// that doesn't start with a space, e.g:
	//	diff --git a/cmd/apply.go b/cmd/apply.go
	//	index b4c2d30..a5f6e5f 100644
	ColorDiffOtherNonSpace = color.New(color.FgHiBlack, color.Italic)
)

func ColorizeDiff(diff string) string {
	if color.NoColor {
		return diff
	}
	lines := strings.Split(diff, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "---"):
			lines[i] = ColorDiffTrippleDash.Sprint(line)
		case strings.HasPrefix(line, "-"):
			lines[i] = ColorDiffRemove.Sprint(line)
		case strings.HasPrefix(line, "+++"):
			lines[i] = ColorDiffTripplePlus.Sprint(line)
		case strings.HasPrefix(line, "+"):
			lines[i] = ColorDiffAdd.Sprint(line)
		case strings.HasPrefix(line, "@@"):
			lines[i] = ColorDiffDoubleAt.Sprint(line)
		case !strings.HasPrefix(line, " "):
			lines[i] = ColorDiffOtherNonSpace.Sprint(line)
		}
	}
	return strings.Join(lines, "\n")
}

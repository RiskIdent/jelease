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

package version

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var parseRegex = regexp.MustCompile(`^(v?)((?:[0-9]+\.?)+)+([a-zA-Z0-9+_.=/-]*)$`)

func Parse(s string) (Version, error) {
	groups := parseRegex.FindStringSubmatch(s)
	if groups == nil {
		return Version{}, errors.New("bad format on version string, expected v0.0.0.0-suffix")
	}
	prefix := groups[1]
	segStrs := strings.Split(groups[2], ".")
	suffix := groups[3]

	segInts := make([]uint, len(segStrs))
	for i, seg := range segStrs {
		parsed, err := strconv.ParseUint(seg, 10, 0)
		if groups == nil {
			return Version{}, fmt.Errorf("version segment: %w", err)
		}
		segInts[i] = uint(parsed)
	}

	return Version{
		Prefix:   prefix,
		Segments: segInts,
		Suffix:   suffix,
	}, nil
}

type Version struct {
	Prefix   string
	Segments []uint
	Suffix   string
}

func (v Version) String() string {
	buf := make([]byte, 0, len(v.Prefix)+len(v.Segments)*3+len(v.Suffix))
	buf = append(buf, []byte(v.Prefix)...)
	for i, seg := range v.Segments {
		if i > 0 {
			buf = append(buf, '.')
		}
		buf = strconv.AppendUint(buf, uint64(seg), 10)
	}
	buf = append(buf, []byte(v.Suffix)...)
	return string(buf)
}

func (v Version) Bump(other Version) Version {
	numSegments := max(len(v.Segments), len(other.Segments))
	segments := make([]uint, numSegments)
	var resetFollowing bool
	for i := 0; i < numSegments; i++ {
		add := indexOrZero(other.Segments, i)
		switch {
		case resetFollowing:
			segments[i] = add
		case add > 0:
			resetFollowing = true
			fallthrough
		default:
			segments[i] = indexOrZero(v.Segments, i) + add
		}
	}
	return Version{
		Prefix:   other.Prefix,
		Segments: segments,
		Suffix:   other.Suffix,
	}
}

func indexOrZero(slice []uint, index int) uint {
	if index < 0 || index >= len(slice) {
		return 0
	}
	return slice[index]
}

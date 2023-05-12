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

package server

import (
	"bytes"
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ginLogger struct {
	defaultLevel zerolog.Level
}

// Ensure it implements the interface
var _ io.Writer = ginLogger{}

func (g ginLogger) Write(b []byte) (int, error) {
	b = bytes.TrimSuffix(b, []byte("\n"))
	level := g.defaultLevel

	var ok bool
	if b, ok = bytes.CutPrefix(b, []byte("[GIN-debug] ")); ok {
		level = zerolog.DebugLevel
	} else {
		b = bytes.TrimPrefix(b, []byte("[GIN] "))
	}
	if b, ok = bytes.CutPrefix(b, []byte("[WARNING] ")); ok {
		level = zerolog.WarnLevel
	}

	log.WithLevel(level).Msgf("[GIN] %s", b)
	return len(b), nil
}

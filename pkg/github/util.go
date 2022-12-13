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

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *client) doRequestUnmarshal(req *http.Request, ptr any) (*http.Response, error) {
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, ptr); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("non-2xx status code: %s", resp.Status)
	}
	return resp, nil
}

func (c *client) newPostJSON(url string, body any) (*http.Request, error) {
	return c.newRequestJSON(http.MethodPost, url, body)
}

func (c *client) newRequestJSON(method, url string, body any) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(method, c.baseURL+url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (c *client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.baseURL+url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("User-Agent", "jelease")
	if c.cred != nil {
		if err := c.cred.Apply(req); err != nil {
			return nil, fmt.Errorf("apply credentials to request: %w", err)
		}
	}
	return req, nil
}

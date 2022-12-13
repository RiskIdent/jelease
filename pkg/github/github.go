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
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client interface {
	CreatePullRequest(repo RepoRef, pr CreatePullRequest) (*PullRequest, error)
}

type RepoRef struct {
	OwnerName string
	RepoName  string
}

// CreatePullRequest hold fields for creating a new pull request.
//
// [https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#create-a-pull-request]
type CreatePullRequest struct {
	Head string `json:"head"`
	Base string `json:"base"`

	Title string `json:"title"`
	Body  string `json:"body"`

	MaintainerCanModify bool `json:"maintainer_can_modify"`
	Draft               bool `json:"draft"`
	Issue               int  `json:"issue"`
}

type PullRequest struct {
	URL               string  `json:"url"`
	ID                int     `json:"id"`
	NodeID            string  `json:"node_id"`
	HTMLURL           string  `json:"html_url"`
	DiffURL           string  `json:"diff_url"`
	PatchURL          string  `json:"patch_url"`
	IssueURL          string  `json:"issue_url"`
	CommitsURL        string  `json:"commits_url"`
	ReviewCommentsURL string  `json:"review_comments_url"`
	ReviewCommentURL  string  `json:"review_comment_url"`
	CommentsURL       string  `json:"comments_url"`
	StatusesURL       string  `json:"statuses_url"`
	Number            int     `json:"number"`
	State             string  `json:"state"`
	Locked            bool    `json:"locked"`
	Title             string  `json:"title"`
	Body              string  `json:"body"`
	Labels            []Label `json:"labels"`
	User              User    `json:"user"`
	ActiveLockReason  string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ClosedAt          time.Time
	MergedAt          time.Time
	MergeCommitSHA    string
	Assignee          User `json:"assignee"`
}

type Label struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
}

type User struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarURL       string `json:"gravatar_url"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type client struct {
	baseURL string
	http    *http.Client
	cred    Credentials
}

func New(baseURL string, credentials Credentials) Client {
	return &client{
		baseURL: baseURL,
		http:    http.DefaultClient,
		cred:    credentials,
	}
}

func (c *client) CreatePullRequest(repo RepoRef, create CreatePullRequest) (*PullRequest, error) {
	req, err := c.newPostJSON(fmt.Sprintf("/repos/%s/%s/pulls",
		url.PathEscape(repo.OwnerName),
		url.PathEscape(repo.RepoName),
	), create)
	if err != nil {
		return nil, err
	}
	var pr PullRequest
	if _, err := c.doRequestUnmarshal(req, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

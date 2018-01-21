/*
   Copyright (C) 2016 The BlameWarrior Authors.

   This file is a part of BlameWarrior service.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package github

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/blamewarrior/collaborators/blamewarrior"
	gh "github.com/google/go-github/github"
)

var (
	ErrRateLimitReached = errors.New("GitHub API request rate limit reached")
	ErrNoSuchRepository = errors.New("no such repository")
)

// Client wraps github.com/google/go-github/github.Client providing methods adapted
// for BlameWarrior use cases.
type Client struct {
	// BaseURL overrides GitHub API endpoint and is intended for use in tests.
	BaseURL *url.URL

	httpClient *http.Client
}

// NewClient returns a new copy of github.Client that uses given http.Client
// to make GitHub API requests.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// RepositoryCollaborators returns GitHub nicknames of collaborators of given
// repository.
func (c *Client) RepositoryCollaborators(repo blamewarrior.Repository) (collaborators []blamewarrior.Account, err error) {
	owner, name := SplitRepositoryName(repo.FullName)

	ctx := context.Background()
	if c.httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: repo.Token})
	httpClient := oauth2.NewClient(ctx, tokenSource)

	api := gh.NewClient(httpClient)
	if c.BaseURL != nil {
		api.BaseURL = c.BaseURL
	}

	opt := &gh.ListOptions{PerPage: 100}
	for {
		users, resp, err := api.Repositories.ListCollaborators(owner, name, opt)
		if err != nil {
			switch err.(type) {
			case *gh.RateLimitError:
				return nil, ErrRateLimitReached
			case *gh.ErrorResponse:
				apiErr := err.(*gh.ErrorResponse)
				if apiErr.Response.StatusCode == http.StatusNotFound {
					return nil, ErrNoSuchRepository
				}
			}

			return nil, fmt.Errorf("request failed: %s", err)
		}

		for _, user := range users {
			if user == nil || user.Login == nil {
				continue
			}
			collaborators = append(collaborators, blamewarrior.Account{
				Nickname: *user.Login,
			},
			)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return collaborators, nil
}

// SplitRepositoryName splits full GitHub repository name into owner and name parts.
func SplitRepositoryName(fullName string) (owner, repo string) {
	sep := strings.IndexByte(fullName, '/')
	if sep <= 0 || sep == len(fullName)-1 {
		return "", ""
	}

	return fullName[0:sep], fullName[sep+1:]
}

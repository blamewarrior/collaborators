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
package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/blamewarrior/collaborators/blamewarrior"
	"github.com/blamewarrior/collaborators/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type tokenServiceMock struct {
	mock.Mock
}

func (tsMock *tokenServiceMock) GetToken(nickname string) (string, error) {
	args := tsMock.Called(nickname)
	return args.String(0), args.Error(1)

}

func TestClient_RepositoryCollaborators(t *testing.T) {
	baseURL, mux, teardown := setupAPIServer()
	defer teardown()

	ts := new(tokenServiceMock)
	ts.On("GetToken", "user1").Return("token1", nil)

	c := github.NewClient(ts)

	ctx := github.Context{context.Background(), baseURL}

	mux.HandleFunc("/repos/user1/repo1/collaborators", func(w http.ResponseWriter, req *http.Request) {
		url := baseURL.String() + "/" + req.URL.Path
		w.Header().Set("Link", `<`+url+`?page=2>; rel="last"`)

		assert.Equal(t, "Bearer token1", req.Header.Get("Authorization"))

		if req.FormValue("page") != "2" {
			w.Header().Set("Link", `<`+url+`?page=2>; rel="next", `+w.Header().Get("Link"))
			w.Write([]byte(`[{"login":"user1", "id": 1, "permissions": {"pull": true,"push": true,"admin": false}},{"login":"user2", "id": 2, "permissions": {"pull": true,"push": true,"admin": false}}]`))
		} else {
			w.Write([]byte(`[{"login":"user3", "id": 3, "permissions": {"pull": true,"push": true,"admin": false}}]`))
		}
	})

	collaborators, err := c.RepositoryCollaborators(ctx, "user1/repo1")
	require.NoError(t, err)
	assert.Len(t, collaborators, 3)
	assert.Contains(t, collaborators, blamewarrior.Account{Login: "user1", Uid: 1, Permissions: map[string]bool{"pull": true, "push": true, "admin": false}})
	assert.Contains(t, collaborators, blamewarrior.Account{Login: "user2", Uid: 2, Permissions: map[string]bool{"pull": true, "push": true, "admin": false}})
	assert.Contains(t, collaborators, blamewarrior.Account{Login: "user3", Uid: 3, Permissions: map[string]bool{"pull": true, "push": true, "admin": false}})

	ts.AssertExpectations(t)
}

func TestClient_RepositoryCollaborators_RepositoryDoesNotExist(t *testing.T) {
	baseURL, mux, teardown := setupAPIServer()
	defer teardown()

	ts := new(tokenServiceMock)
	ts.On("GetToken", "user1").Return("token1", nil)

	c := github.NewClient(ts)

	ctx := github.Context{context.Background(), baseURL}

	mux.HandleFunc("/repos/user1/repo1/collaborators", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	})

	_, err := c.RepositoryCollaborators(ctx, "user1/repo1")
	assert.Equal(t, github.ErrNoSuchRepository, err)

	ts.AssertExpectations(t)
}

func TestClient_RepositoryCollaborators_RateLimitReached(t *testing.T) {
	baseURL, mux, teardown := setupAPIServer()
	defer teardown()

	ts := new(tokenServiceMock)
	ts.On("GetToken", "user1").Return("token1", nil)

	c := github.NewClient(ts)

	ctx := github.Context{context.Background(), baseURL}

	mux.HandleFunc("/repos/user1/repo1/collaborators", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Unix(), 10))
		http.Error(w, `{"message":"API rate limit exceeded for 127.0.0.1"}`, http.StatusForbidden)
	})

	_, err := c.RepositoryCollaborators(ctx, "user1/repo1")
	require.Error(t, err)
	assert.Equal(t, github.ErrRateLimitReached, err)

	ts.AssertExpectations(t)
}

func TestSplitRepositoryName(t *testing.T) {
	examples := map[string]struct {
		Owner, Name string
	}{
		"":     {"", ""},
		"a":    {"", ""},
		"a/":   {"", ""},
		"/b":   {"", ""},
		"a/b":  {"a", "b"},
		"a/b/": {"a", "b/"},
		"a//b": {"a", "/b"},
	}

	for fullName, expected := range examples {
		t.Run(fmt.Sprintf("fullName: %q", fullName), func(t *testing.T) {
			owner, name := github.SplitRepositoryName(fullName)
			assert.Equal(t, expected.Owner, owner)
			assert.Equal(t, expected.Name, name)
		})
	}
}

func setupAPIServer() (baseURL *url.URL, mux *http.ServeMux, teardownFn func()) {
	mux = http.NewServeMux()
	srv := httptest.NewServer(mux)
	baseURL, _ = url.Parse(srv.URL)

	return baseURL, mux, srv.Close
}

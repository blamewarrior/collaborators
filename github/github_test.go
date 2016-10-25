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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/blamewarrior/collaborators/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_RepositoryCollaborators(t *testing.T) {
	baseURL, mux, teardown := setupAPIServer()
	defer teardown()

	c := github.NewClient(nil)
	c.SetBaseURL(baseURL)

	mux.HandleFunc("/repos/user1/repo1/collaborators", func(w http.ResponseWriter, req *http.Request) {
		url := baseURL.String() + "/" + req.URL.Path
		w.Header().Set("Link", `<`+url+`?page=2>; rel="last"`)

		if req.FormValue("page") != "2" {
			w.Header().Set("Link", `<`+url+`?page=2>; rel="next", `+w.Header().Get("Link"))
			w.Write([]byte(`[{"login":"user1"},{"login":"user2"}]`))
		} else {
			w.Write([]byte(`[{"login":"user3"}]`))
		}
	})

	collaborators, err := c.RepositoryCollaborators("user1/repo1")
	require.NoError(t, err)
	assert.Len(t, collaborators, 3)
	assert.Contains(t, collaborators, "user1")
	assert.Contains(t, collaborators, "user2")
	assert.Contains(t, collaborators, "user3")
}

func TestClient_RepositoryCollaborators_RepositoryDoesNotExist(t *testing.T) {
	baseURL, mux, teardown := setupAPIServer()
	defer teardown()

	c := github.NewClient(nil)
	c.SetBaseURL(baseURL)

	mux.HandleFunc("/repos/user1/repo1/collaborators", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	})

	_, err := c.RepositoryCollaborators("user1/repo1")
	assert.Equal(t, github.ErrNoSuchRepository, err)
}

func TestClient_RepositoryCollaborators_RateLimitReached(t *testing.T) {
	baseURL, mux, teardown := setupAPIServer()
	defer teardown()

	c := github.NewClient(nil)
	c.SetBaseURL(baseURL)

	mux.HandleFunc("/repos/user1/repo1/collaborators", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Unix(), 10))
		http.Error(w, `{"message":"API rate limit exceeded for 127.0.0.1"}`, http.StatusForbidden)
	})

	_, err := c.RepositoryCollaborators("user1/repo1")
	assert.Equal(t, github.ErrRateLimitReached, err)
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

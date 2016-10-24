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
	"testing"

	"github.com/blamewarrior/collaborators/github"
	"github.com/stretchr/testify/assert"
)

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

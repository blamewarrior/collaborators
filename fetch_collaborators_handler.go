/*
   Copyright (C) 2017 The BlameWarrior Authors.
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

package main

import (
	"fmt"
	"net/http"

	"github.com/blamewarrior/collaborators/blamewarrior"
	"github.com/blamewarrior/collaborators/blamewarrior/tokens"
)

type FetchCollaboratorsHandler struct {
	hostname      string
	collaboration blamewarrior.Collaboration
	tokenClient   tokens.Client
}

func (h *FetchCollaboratorsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	username := req.URL.Query().Get(":username")
	repo := req.URL.Query().Get(":repo")

	fullName := fmt.Sprintf("%s/%s", username, repo)

	fmt.Println(fullName)
}

func NewFetchCollaboratorsHandler(hostname string, collaboration blamewarrior.Collaboration,
	tokenClient tokens.Client) *FetchCollaboratorsHandler {

	return &FetchCollaboratorsHandler{
		hostname:      hostname,
		collaboration: collaboration,
		tokenClient:   tokenClient,
	}
}

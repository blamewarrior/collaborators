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
	"database/sql"
	"fmt"
	"net/http"

	"github.com/blamewarrior/collaborators/blamewarrior"
)

type EditCollaboratorHandler struct {
	hostname      string
	db            *sql.DB
	collaboration blamewarrior.Collaboration
}

func (h *EditCollaboratorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	username := req.URL.Query().Get(":username")
	repo := req.URL.Query().Get(":repo")

	fullName := fmt.Sprintf("%s/%s", username, repo)

	fmt.Println(fullName)
}

func NewEditCollaboratorHandler(hostname string, db *sql.DB, collaboration blamewarrior.Collaboration) *EditCollaboratorHandler {
	return &EditCollaboratorHandler{
		hostname:      hostname,
		db:            db,
		collaboration: collaboration,
	}
}

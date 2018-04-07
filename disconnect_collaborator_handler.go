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
	"log"
	"net/http"

	"github.com/blamewarrior/collaborators/blamewarrior"
)

type DisconnectCollaboratorHandler struct {
	hostname      string
	db            *sql.DB
	collaboration blamewarrior.Collaboration
}

func (h *DisconnectCollaboratorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	username := req.URL.Query().Get(":username")
	repo := req.URL.Query().Get(":repo")

	collaboratorName := req.URL.Query().Get(":collaborator")

	fullName := fmt.Sprintf("%s/%s", username, repo)

	if username == "" || repo == "" {
		http.Error(w, "Incorrect full name", http.StatusBadRequest)
		return
	}

	if collaboratorName == "" {
		http.Error(w, "Incorrect collaborator name", http.StatusBadRequest)
		return
	}

	if err := h.collaboration.DisconnectAccount(h.db, fullName, collaboratorName); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func NewDisconnectCollaboratorHandler(hostname string, db *sql.DB, collaboration blamewarrior.Collaboration) *DisconnectCollaboratorHandler {
	return &DisconnectCollaboratorHandler{
		hostname:      hostname,
		db:            db,
		collaboration: collaboration,
	}
}

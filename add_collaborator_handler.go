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
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/blamewarrior/collaborators/blamewarrior"
	bw "github.com/blamewarrior/collaborators/blamewarrior"
)

type AddCollaboratorHandler struct {
	hostname      string
	db            *sql.DB
	collaboration blamewarrior.Collaboration
}

func (h *AddCollaboratorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	username := req.URL.Query().Get(":username")
	repo := req.URL.Query().Get(":repo")

	if username == "" || repo == "" {
		http.Error(w, "Incorrect full name", http.StatusBadRequest)
		return
	}

	fullName := fmt.Sprintf("%s/%s", username, repo)

	var account bw.Account

	if err := json.NewDecoder(req.Body).Decode(&account); err != nil {
		http.Error(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	if err := h.AddCollaborator(fullName, &account); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AddCollaboratorHandler) AddCollaborator(fullName string, account *bw.Account) (err error) {
	tx, err := h.db.Begin()

	if err != nil {
		return err
	}

	_, err = h.collaboration.AddAccount(tx, fullName, account)

	return err
}

func NewAddCollaboratorHandler(hostname string, db *sql.DB, collaboration blamewarrior.Collaboration) *AddCollaboratorHandler {
	return &AddCollaboratorHandler{
		hostname:      hostname,
		db:            db,
		collaboration: collaboration,
	}
}

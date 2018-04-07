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
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/blamewarrior/collaborators/blamewarrior"
	"github.com/blamewarrior/collaborators/github"
)

type FetchCollaboratorsHandler struct {
	hostname      string
	db            *sql.DB
	collaboration blamewarrior.Collaboration
	githubClient  *github.Client

	GithubBaseURL *url.URL
}

func (h *FetchCollaboratorsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	username := req.URL.Query().Get(":username")
	repo := req.URL.Query().Get(":repo")

	if username == "" || repo == "" {
		http.Error(w, "Incorrect full name", http.StatusBadRequest)
		return
	}

	fullName := fmt.Sprintf("%s/%s", username, repo)

	err := h.fetchCollaborators(fullName)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
	}

}

func (h *FetchCollaboratorsHandler) fetchCollaborators(fullName string) error {
	var ctx github.Context

	ctx = github.Context{context.Background(), h.GithubBaseURL}

	collaborators, err := h.githubClient.RepositoryCollaborators(ctx, fullName)

	if err != nil {
		return err
	}

	tx, err := h.db.Begin()

	defer tx.Rollback()

	if err != nil {
		return err
	}

	if err := h.collaboration.CreateRepository(tx, fullName); err != nil {
		return err
	}

	for _, collaborator := range collaborators {

		account := &blamewarrior.Account{
			Uid:         collaborator.Uid,
			Login:       collaborator.Login,
			Permissions: collaborator.Permissions,
		}

		_, err := h.collaboration.AddAccount(tx, fullName, account)

		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func NewFetchCollaboratorsHandler(hostname string, db *sql.DB, collaboration blamewarrior.Collaboration,
	githubClient *github.Client) *FetchCollaboratorsHandler {

	return &FetchCollaboratorsHandler{
		hostname:      hostname,
		db:            db,
		collaboration: collaboration,
		githubClient:  githubClient,
	}
}

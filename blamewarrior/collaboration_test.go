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
package blamewarrior_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/blamewarrior/collaborators/blamewarrior"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryAddAccount(t *testing.T) {

	results := []struct {
		Account *blamewarrior.Account
		Err     error
	}{
		{
			Account: &blamewarrior.Account{
				Uid:         123,
				Login:       "octocat",
				Permissions: map[string]bool{"admin": true},
			},
			Err: nil,
		},
	}

	for _, result := range results {
		db, teardown := setup()

		var repositoryId int

		err := db.QueryRow(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos").Scan(&repositoryId)
		require.NoError(t, err)

		account := result.Account
		repositoriesService := blamewarrior.NewCollaborationService()

		account, err = repositoriesService.AddAccount(db, "blamewarrior/repos", account)
		assert.Equal(t, result.Err, err)
		assert.NotEmpty(t, account.Id)

		var obtainedRepositoryId, obtainedAccountId int
		err = db.QueryRow("SELECT repository_id FROM collaboration").Scan(&obtainedRepositoryId)
		require.NoError(t, err)
		require.Equal(t, repositoryId, obtainedRepositoryId)

		err = db.QueryRow("SELECT account_id FROM collaboration").Scan(&obtainedAccountId)
		require.NoError(t, err)
		require.Equal(t, account.Id, obtainedAccountId)

		teardown()
	}
}

func TestRepositoryListAccount(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	var accountId int
	_, err := db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos")
	require.NoError(t, err)
	err = db.QueryRow(blamewarrior.AddAccountQuery, 123, "octocat", `{"admin": true}`).Scan(&accountId)
	require.NoError(t, err)
	_, err = db.Exec(blamewarrior.BuildCollaborationQuery, "blamewarrior/repos", accountId)
	require.NoError(t, err)

	repositoriesService := blamewarrior.NewCollaborationService()
	accounts, err := repositoriesService.ListAccounts(db, "blamewarrior/repos")

	require.NoError(t, err)
	assert.NotEmpty(t, accounts)
}

func TestRepositoryDisconnectAccount(t *testing.T) {
	db, teardown := setup()
	defer teardown()
	var blamewarriorReposId, blamewarriorHooksId int
	var octocatId, octocatTstId int

	err := db.QueryRow(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos").Scan(&blamewarriorReposId)
	require.NoError(t, err)
	err = db.QueryRow(blamewarrior.CreateRepositoryQuery, "blamewarrior/hooks").Scan(&blamewarriorHooksId)
	require.NoError(t, err)

	err = db.QueryRow(blamewarrior.AddAccountQuery, 123, "octocat", "{}").Scan(&octocatId)
	require.NoError(t, err)

	err = db.QueryRow(blamewarrior.AddAccountQuery, 1234, "octocat_tst", "{}").Scan(&octocatTstId)
	require.NoError(t, err)

	_, err = db.Exec(blamewarrior.BuildCollaborationQuery, "blamewarrior/repos", octocatId)
	require.NoError(t, err)

	_, err = db.Exec(blamewarrior.BuildCollaborationQuery, "blamewarrior/hooks", octocatTstId)
	require.NoError(t, err)

	repositoriesService := blamewarrior.NewCollaborationService()
	err = repositoriesService.DisconnectAccount(db, "blamewarrior/repos", "octocat")
	require.NoError(t, err)

	var collaborationCount int
	err = db.QueryRow("SELECT COUNT(*) FROM collaboration where repository_id = $1", blamewarriorReposId).Scan(&collaborationCount)
	require.NoError(t, err)
	assert.Zero(t, collaborationCount)

	var obtainedRepositoryId, obtainedAccountId int

	err = db.QueryRow("SELECT repository_id, account_id FROM collaboration").Scan(&obtainedRepositoryId, &obtainedAccountId)
	require.NoError(t, err)
	assert.Equal(t, blamewarriorHooksId, obtainedRepositoryId)
	assert.Equal(t, octocatTstId, obtainedAccountId)
}

func TestRepositoryEditAccount(t *testing.T) {

	results := []struct {
		Account *blamewarrior.Account
		Err     error
	}{
		{
			Account: &blamewarrior.Account{
				Uid:         126,
				Login:       "octocat_client",
				Permissions: map[string]bool{"admin": true},
			},
			Err: nil,
		},
	}

	for _, result := range results {
		db, teardown := setup()

		var accountId int
		_, err := db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos")
		require.NoError(t, err)
		err = db.QueryRow(blamewarrior.AddAccountQuery, result.Account.Uid, result.Account.Login, "{}").Scan(&accountId)
		require.NoError(t, err)
		_, err = db.Exec(blamewarrior.BuildCollaborationQuery, "blamewarrior/repos", accountId)
		require.NoError(t, err)

		account := result.Account
		repositoriesService := blamewarrior.NewCollaborationService()
		err = repositoriesService.EditAccount(db, "blamewarrior/repos", account)
		assert.Equal(t, result.Err, err)
		teardown()
	}
}

func setup() (tx *sql.Tx, teardownFn func()) {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("missing test database name (expected to be passed via ENV['DB_NAME'])")
	}

	opts := &blamewarrior.DatabaseOptions{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	db, err := blamewarrior.ConnectDatabase(dbName, opts)
	if err != nil {
		log.Fatalf("failed to establish connection with test db %s using connection string %s: %s", dbName, opts.ConnectionString(), err)
	}

	tx, err = db.Begin()

	if err != nil {
		log.Fatalf("failed to create transaction, %s", err)
	}

	return tx, func() {
		tx.Rollback()
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}
}

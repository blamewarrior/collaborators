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

	_ "github.com/lib/pq"
)

func TestGetRepositories_FetchRepositories(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")
	require.NoError(t, err)

	firstUserID, err := createUser("user1", "token1", db)
	require.NoError(t, err)
	require.NoError(t, createRepository(firstUserID, "user1/repo1", true, db))

	secondUserID, err := createUser("user2", "token2", db)
	require.NoError(t, err)
	require.NoError(t, createRepository(secondUserID, "user2/repo2", true, db))

	repos, err := blamewarrior.GetRepositories(db)
	require.NoError(t, err)

	assert.Len(t, repos, 2)
	assert.Contains(t, repos, blamewarrior.Repository{FullName: "user1/repo1", Token: "token1"})
	assert.Contains(t, repos, blamewarrior.Repository{FullName: "user2/repo2", Token: "token2"})
}

func TestGetRepositories_OnlyTracked(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")
	require.NoError(t, err)

	userID, err := createUser("user", "token", db)
	require.NoError(t, err)

	require.NoError(t, createRepository(userID, "user/tracked1", true, db))
	require.NoError(t, createRepository(userID, "user/not_tracked", false, db))
	require.NoError(t, createRepository(userID, "user/tracked2", true, db))

	repos, err := blamewarrior.GetRepositories(db)
	require.NoError(t, err)

	assert.Len(t, repos, 2)
	assert.NotContains(t, repos, blamewarrior.Repository{FullName: "user/not_tracked"})
}

func setup() (db *sql.DB, teardownFn func()) {
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

	return db, func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}
}

func createUser(user, token string, db *sql.DB) (userId int64, err error) {
	err = db.QueryRow("INSERT INTO github_accounts (nickname, token) VALUES ($1, $2) RETURNING id;", user, token).Scan(&userId)
	return userId, err
}

func createRepository(userId int64, fullName string, tracked bool, db *sql.DB) error {
	_, err := db.Exec("INSERT INTO repositories (github_account_id, full_name, tracked) VALUES ($1, $2, $3);", userId, fullName, tracked)
	return err
}

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
	"strconv"
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

	r, err := db.Exec("INSERT INTO repositories (full_name, tracked) VALUES ($1, $2), ($3, $4);",
		"user1/repo1", true,
		"user2/repo2", true,
	)
	require.NoError(t, err)

	insertedNum, err := r.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 2, insertedNum)

	repos, err := blamewarrior.GetRepositories(db)
	require.NoError(t, err)

	assert.Len(t, repos, 2)
	assert.Contains(t, repos, blamewarrior.Repository{FullName: "user1/repo1"})
	assert.Contains(t, repos, blamewarrior.Repository{FullName: "user2/repo2"})
}

func TestGetRepositories_OnlyTracked(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")
	require.NoError(t, err)

	r, err := db.Exec("INSERT INTO repositories (full_name, tracked) VALUES ($1, $2), ($3, $4), ($5, $6);",
		"user/tracked1", true,
		"user/not_tracked", false,
		"user/tracked2", true,
	)
	require.NoError(t, err)

	insertedNum, err := r.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 3, insertedNum)

	repos, err := blamewarrior.GetRepositories(db)
	require.NoError(t, err)

	assert.Len(t, repos, 2)
	assert.NotContains(t, repos, blamewarrior.Repository{FullName: "user/not_tracked"})
}

func setup() (db *sql.DB, teardownFn func()) {
	connStr := "sslmode=disable"

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		connStr += " dbname=" + dbName
	} else {
		log.Fatal("missing test database name (expected to be passed via ENV['DB_NAME'])")
	}

	if user := os.Getenv("DB_USER"); user != "" {
		connStr += " user=" + user
	}

	if password := os.Getenv("DB_PASSWORD"); password != "" {
		connStr += " password=" + strconv.Quote(password)
	}

	if host := os.Getenv("DB_HOST"); host != "" {
		connStr += " host=" + host
	}

	if port := os.Getenv("DB_PORT"); port != "" {
		connStr += " port=" + port
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to db using connection string %q", connStr)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}
}

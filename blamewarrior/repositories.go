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
package blamewarrior

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Repository represents a single GitHub repository connected to BlameWarrior.
type Repository struct {
	// FullName is the name of GitHub repository preceded with the name of
	// the owner, e.g. blamewarrior/collaborators.
	FullName string
}

// GetRepository is a method to fetch all tracked repositories from database.
func GetRepositories(db *sql.DB) (repos []Repository, err error) {
	rows, err := db.Query("SELECT full_name FROM repositories WHERE tracked;")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var repo Repository
		if err = rows.Scan(&repo.FullName); err != nil {
			return nil, fmt.Errorf("failed to fetch repository: %s", err)
		}

		repos = append(repos, repo)
	}

	return repos, rows.Err()
}

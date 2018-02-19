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
	"encoding/json"
	"fmt"
)

// Account represents GitHub user account stored in BlameWarrior database.
type Account struct {
	Id         int             `json:"-"`
	Uid        int             `json:"uid"`
	Login      string          `json:"login"`
	Permisions json.RawMessage `json:"permissions"`
}

type Accounts interface {
	List(sqlRunner SQLRunner, repositoryFullName string) ([]Account, error)
	Add(sqlRunner SQLRunner, repositoryFullName string, account *Account) (*Account, error)
	Edit(sqlRunner SQLRunner, repositoryFullName string, account *Account) error
	Delete(sqlRunner SQLRunner, repositoryFullName string, login string) error
}

type AccountsService struct {
	GithubRepoName string
}

func (repo *AccountsService) List(sqlRunner SQLRunner, repositoryFullName string) ([]Account, error) {
	accounts := make([]Account, 0)
	rows, err := sqlRunner.Query(GetListAccountsQuery, repositoryFullName)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		account := Account{}

		if err := rows.Scan(
			&account.Id,
			&account.Uid,
			&account.Login,
			&account.Permisions,
		); err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}
func (repo *AccountsService) Add(tx *sql.Tx, repositoryFullName string, account *Account) (*Account, error) {

	err := tx.QueryRow(AddAccountQuery,
		account.Uid,
		account.Login,
		account.Permisions,
	).Scan(&account.Id)

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %s", err)
	}

	_, err = tx.Exec(BuildCollaborationQuery,
		repositoryFullName,
		account.Id,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create collaboration: %s", err)
	}

	return account, nil
}
func (repo *AccountsService) Edit(sqlRunner SQLRunner, repositoryFullName string, account *Account) error {
	_, err := sqlRunner.Exec(EditAccountQuery,
		account.Id,
		account.Uid,
		account.Login,
		account.Permisions,
	)

	if err != nil {
		return fmt.Errorf("failed to update account: %s", err)
	}

	return err
}
func (repo *AccountsService) Delete(sqlRunner SQLRunner, repositoryFullName, login string) error {
	if _, err := sqlRunner.Exec(DeleteAccountQuery, repositoryFullName, login); err != nil {
		return fmt.Errorf("failed to delete account: %s", err)
	}

	return nil
}

const (
	GetListAccountsQuery = `
     SELECT accounts.id, accounts.uid, accounts.login, accounts.permissions
         FROM accounts
         INNER JOIN collaboration ON accounts.id = collaboration.account_id
         INNER JOIN repositories ON collaboration.repository_id = repositories.id
         WHERE repositories.full_name = $1
   `
	AddAccountQuery = `
      INSERT INTO accounts(uid, login, permissions) VALUES ($1, $2, $3) RETURNING id
  `

	BuildCollaborationQuery = `
    INSERT INTO collaboration (SELECT DISTINCT id, $2::int FROM repositories WHERE full_name=$1)
  `

	EditAccountQuery = `
      UPDATE accounts SET uid=$2, login=$3, permissions=$4 WHERE id=$1;
   `

	DeleteAccountQuery = `
      WITH account AS (
        SELECT id FROM accounts WHERE login=$2
      )
      DELETE FROM collaboration WHERE account_id =(
        SELECT id from account
      ) AND repository_id = (SELECT id FROM repositories WHERE full_name = $1 LIMIT 1)
   `
)

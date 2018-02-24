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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type AccountPermissions map[string]bool

func (perms AccountPermissions) Value() (driver.Value, error) {
	b, err := json.Marshal(perms)
	return b, err
}

func (perms *AccountPermissions) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	err := json.Unmarshal(source, &perms)
	if err != nil {
		return err
	}

	return nil
}

// Account represents GitHub user account stored in BlameWarrior database.
type Account struct {
	Id          int                `json:"-"`
	Uid         int                `json:"uid"`
	Login       string             `json:"login"`
	Permissions AccountPermissions `json:"permissions"`
}

type Collaboration interface {
	CreateRepository(sqlRunner SQLRunner, repositoryFullName string) error
	ListAccounts(sqlRunner SQLRunner, repositoryFullName string) ([]Account, error)
	AddAccount(tx *sql.Tx, repositoryFullName string, account *Account) (*Account, error)
	EditAccount(sqlRunner SQLRunner, account *Account) error
	DisconnectAccount(sqlRunner SQLRunner, repositoryFullName, login string) error
}

type CollaborationService struct{}

func NewCollaborationService() *CollaborationService {
	return new(CollaborationService)
}

func (service *CollaborationService) CreateRepository(sqlRunner SQLRunner, repositoryFullName string) error {
	_, err := sqlRunner.Exec(CreateRepositoryQuery, repositoryFullName)
	return err
}

func (service *CollaborationService) ListAccounts(sqlRunner SQLRunner, repositoryFullName string) ([]Account, error) {
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
			&account.Permissions,
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
func (service *CollaborationService) AddAccount(tx *sql.Tx, repositoryFullName string, account *Account) (*Account, error) {
	err := tx.QueryRow("SELECT id FROM accounts WHERE login = $1 LIMIT 1", account.Login).Scan(&account.Id)

	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}

		if err = tx.QueryRow(AddAccountQuery,
			account.Uid,
			account.Login,
			account.Permissions,
		).Scan(&account.Id); err != nil {
			return nil, fmt.Errorf("failed to create account: %s", err)
		}
	}

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

func (service *CollaborationService) EditAccount(sqlRunner SQLRunner, account *Account) error {
	_, err := sqlRunner.Exec(EditAccountQuery,
		account.Id,
		account.Uid,
		account.Login,
		account.Permissions,
	)

	if err != nil {
		return fmt.Errorf("failed to update account: %s", err)
	}

	return err
}
func (service *CollaborationService) DisconnectAccount(sqlRunner SQLRunner, repositoryFullName, login string) error {
	if _, err := sqlRunner.Exec(DisconnectAccountQuery, repositoryFullName, login); err != nil {
		return fmt.Errorf("failed to delete account: %s", err)
	}

	return nil
}

const (
	CreateRepositoryQuery = `
    INSERT INTO repositories(full_name) VALUES($1) RETURNING id
  `

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

	DisconnectAccountQuery = `
      WITH account AS (
        SELECT id FROM accounts WHERE login=$2
      )
      DELETE FROM collaboration WHERE account_id =(
        SELECT id from account
      ) AND repository_id = (SELECT id FROM repositories WHERE full_name = $1 LIMIT 1)
   `
)

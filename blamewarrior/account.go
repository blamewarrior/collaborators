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
	"encoding/json"
	"fmt"
)

// Account represents GitHub user account stored in BlameWarrior database.
type Account struct {
	Id         int             `json:"-"`
	Uid        int             `json:"uid"`
	Login      string          `json:"login"`
	Repository string          `json:"repository"`
	Permisions json.RawMessage `json:"permissions"`
}

type Accounts interface {
	List(sqlRunner SQLRunner, repositoryFullName string) ([]Account, error)
	Add(sqlRunner SQLRunner, account *Account) (*Account, error)
	Edit(sqlRunner SQLRunner, account *Account) error
	Delete(sqlRunner SQLRunner, login string) error
}

type AccountsRepository struct {
}

func (repo *AccountsRepository) List(sqlRunner SQLRunner, repositoryFullName string) ([]Account, error) {
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
			&account.Repository,
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
func (repo *AccountsRepository) Add(sqlRunner SQLRunner, account *Account) (*Account, error) {
	err := sqlRunner.QueryRow(AddAcountQuery,
		account.Uid,
		account.Login,
		account.Repository,
		account.Permisions,
	).Scan(&account.Id)

	if err != nil {
		return account, fmt.Errorf("failed to create account: %s", err)
	}

	return account, err
}
func (repo *AccountsRepository) Edit(sqlRunner SQLRunner, account *Account) error {
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
func (repo *AccountsRepository) Delete(sqlRunner SQLRunner, repositoryFullName, login string) error {
	if _, err := sqlRunner.Exec(DeleteAccountQuery, login); err != nil {
		return fmt.Errorf("failed to delete account: %s", err)
	}

	return nil
}

const (
	GetListAccountsQuery = `
      SELECT accounts.id, accounts.uid, accounts.login, accounts.permissions
         FROM accounts INNER JOIN repositories ON accounts.id = repositories.account_id
         WHERE repositories.full_name = $1
   `
	AddAcountQuery = `
      INSERT INTO accounts(uid, login, permissions) VALUES ($1, $2, $3) RETURNING id
   `

	EditAccountQuery = `
      UPDATE accounts SET uid=$2, login=$3, permissions=$4 WHERE id=$1;
   `

	DeleteAccountQuery = `
      DELETE FROM accounts WHERE login=$1
   `
)

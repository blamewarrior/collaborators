package blamewarrior_test

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/blamewarrior/collaborators/blamewarrior"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAccount(t *testing.T) {

	results := []struct {
		Account *blamewarrior.Account
		Err     error
	}{
		{
			Account: &blamewarrior.Account{
				Uid:        123,
				Login:      "octocat",
				Permisions: json.RawMessage("{}"),
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
		accountsService := new(blamewarrior.AccountsService)
		account, err = accountsService.Add(db, "blamewarrior/repos", account)
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

func TestListAccount(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	var accountId int
	_, err := db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos")
	require.NoError(t, err)
	err = db.QueryRow(blamewarrior.AddAccountQuery, 123, "octocat", "{}").Scan(&accountId)
	require.NoError(t, err)
	_, err = db.Exec(blamewarrior.BuildCollaborationQuery, "blamewarrior/repos", accountId)
	require.NoError(t, err)

	accountsService := new(blamewarrior.AccountsService)
	accounts, err := accountsService.List(db, "blamewarrior/repos")
	require.NoError(t, err)
	assert.NotEmpty(t, accounts)
}

// func TestDeleteAccount(t *testing.T) {
// 	db, teardown := setup()
// 	defer teardown()

// 	_, err := db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos")
// 	require.NoError(t, err)
// 	_, err = db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/hooks")
// 	require.NoError(t, err)

// 	_, err = db.Exec(blamewarrior.AddAccountQuery, 123, "octocat", "{}", "blamewarrior/repos")
// 	require.NoError(t, err)

// 	_, err = db.Exec(blamewarrior.AddAccountQuery, 123, "octocat", "{}", "blamewarrior/hooks")
// 	require.NoError(t, err)

// 	accountsService := new(blamewarrior.AccountsService)
// 	err = accountsService.Delete(db, "blamewarrior/repos", "octocat")
// 	require.NoError(t, err)
// }

func TestEditAccount(t *testing.T) {

	results := []struct {
		Account *blamewarrior.Account
		Err     error
	}{
		{
			Account: &blamewarrior.Account{
				Uid:        126,
				Login:      "octocat_client",
				Permisions: json.RawMessage("{}"),
			},
			Err: nil,
		},
	}

	for _, result := range results {
		db, teardown := setup()

		var accountId int
		_, err := db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos")
		require.NoError(t, err)
		err = db.QueryRow(blamewarrior.AddAccountQuery, 123, "octocat", "{}").Scan(&accountId)
		require.NoError(t, err)
		_, err = db.Exec(blamewarrior.BuildCollaborationQuery, "blamewarrior/repos", accountId)
		require.NoError(t, err)

		account := result.Account
		accountsService := new(blamewarrior.AccountsService)
		err = accountsService.Edit(db, "blamewarrior/repos", account)
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

package blamewarrior_test

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/blamewarrior/collaborators/blamewarrior"

	"github.com/stretchr/testify/assert"
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
				Repository: "octocat/test",
				Permisions: json.RawMessage("{}"),
			},
			Err: nil,
		},
	}

	for _, result := range results {
		db, teardown := setup()

		account := result.Account
		accountsRepository := new(blamewarrior.AccountsRepository)
		account, err := accountsRepository.Add(db, account)
		assert.Equal(t, result.Err, err)
		assert.NotEmpty(t, account.Id)
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

package main_test

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/blamewarrior/collaborators/blamewarrior"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/blamewarrior/collaborators"
)

func TestAddCollaboratorHandler(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories, collaboration, accounts;")

	require.NoError(t, err)

	_, err = db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/repos")
	require.NoError(t, err)

	results := []struct {
		Owner        string
		Name         string
		ResponseCode int
		ResponseBody string
	}{
		{
			Owner:        "",
			Name:         "",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: "Incorrect full name\n",
		},
		{
			Owner:        "blamewarrior",
			Name:         "test",
			ResponseCode: http.StatusOK,
			ResponseBody: "",
		},
	}

	for _, result := range results {
		req, err := http.NewRequest("POST", "/repositories?:owner="+result.Owner+"&:name="+result.Name, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		collaboration := blamewarrior.NewCollaborationService()

		handler := main.NewAddCollaboratorHandler("blamewarrior.com", db, collaboration)
		handler.ServeHTTP(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))
	}
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

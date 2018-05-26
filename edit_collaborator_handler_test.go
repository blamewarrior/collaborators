package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blamewarrior/collaborators/blamewarrior"

	main "github.com/blamewarrior/collaborators"
)

func TestEditCollaboratorHandler(t *testing.T) {
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
			Name:         "test_edit_account",
			ResponseCode: http.StatusOK,
			ResponseBody: "",
		},
	}

	for _, result := range results {
		db, teardown := setupTestDBConn()
		_, err := db.Exec("TRUNCATE repositories, collaboration, accounts")
		require.NoError(t, err)

		_, err = db.Exec(blamewarrior.CreateRepositoryQuery, fmt.Sprintf("%s/%s", result.Owner, result.Name))
		require.NoError(t, err)
		req, err := http.NewRequest("POST", "/collaborators?:username="+result.Owner+"&:repo="+result.Name, bytes.NewBufferString(addCollaboratorRequestBody))
		require.NoError(t, err)

		w := httptest.NewRecorder()

		collaboration := blamewarrior.NewCollaborationService()

		handler := main.NewEditCollaboratorHandler("blamewarrior.com", db, collaboration)
		handler.ServeHTTP(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))

		teardown()
	}
}

const (
	editCollaboratorRequestBody = `
    {
      "uid": 1345,
      "login": "blamewarrior",
      "permissions": {
        "admin": false
      }
    }

  `
)

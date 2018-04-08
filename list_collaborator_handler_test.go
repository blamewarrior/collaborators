package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/blamewarrior/collaborators"
	"github.com/blamewarrior/collaborators/blamewarrior"
)

func TestListCollaboratorHandler(t *testing.T) {

	results := []struct {
		Owner        string
		Name         string
		AccountLogin string
		ResponseCode int
		ResponseBody string
	}{
		{
			Owner:        "",
			Name:         "",
			AccountLogin: "",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: "Incorrect full name\n",
		},
		{
			Owner:        "blamewarrior",
			Name:         "test_list_handler",
			AccountLogin: "octocat",
			ResponseCode: http.StatusOK,
			ResponseBody: "[{\"uid\":123,\"login\":\"octocat\",\"permissions\":{\"admin\":true}}]\n",
		},
	}

	for _, result := range results {
		db, teardown := setupTestDBConn()

		var accountId int
		_, err := db.Exec(blamewarrior.CreateRepositoryQuery, fmt.Sprintf("%s/%s", result.Owner, result.Name))
		require.NoError(t, err)
		err = db.QueryRow(blamewarrior.AddAccountQuery, 123, result.AccountLogin, `{"admin": true}`).Scan(&accountId)
		require.NoError(t, err)
		_, err = db.Exec(blamewarrior.BuildCollaborationQuery, fmt.Sprintf("%s/%s", result.Owner, result.Name), accountId)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/collaborators?:username="+result.Owner+"&:repo="+result.Name, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		collaboration := blamewarrior.NewCollaborationService()

		handler := main.NewListCollaboratorHandler("blamewarrior.com", db, collaboration)
		handler.ServeHTTP(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))

		teardown()
	}
}

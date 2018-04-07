package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blamewarrior/collaborators/blamewarrior"
	"github.com/blamewarrior/collaborators/blamewarrior/tokens"
	"github.com/blamewarrior/collaborators/github"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/blamewarrior/collaborators"
)

func TestFetchCollaboratorHandler(t *testing.T) {

	results := []struct {
		Owner         string
		Name          string
		ResponseCode  int
		ResponseBody  string
		Collaborators []blamewarrior.Account
	}{
		{
			Owner:         "",
			Name:          "",
			ResponseCode:  http.StatusBadRequest,
			ResponseBody:  "Incorrect full name\n",
			Collaborators: []blamewarrior.Account{},
		},
		{
			Owner:        "blamewarrior",
			Name:         "test",
			ResponseCode: http.StatusOK,
			ResponseBody: "",
			Collaborators: []blamewarrior.Account{
				blamewarrior.Account{
					Uid:         1,
					Login:       "user1",
					Permissions: blamewarrior.AccountPermissions{"pull": true, "push": true, "admin": false},
				},
			},
		},
	}

	for _, result := range results {
		db, teardownDB := setupTestDBConn()

		_, err := db.Exec("TRUNCATE repositories, collaboration, accounts")

		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/repositories?:username="+result.Owner+"&:repo="+result.Name, bytes.NewBufferString(addCollaboratorRequestBody))
		require.NoError(t, err)

		w := httptest.NewRecorder()

		testAPIEndpoint, mux, teardownAPIServer := setupAPIServer()

		tokenClient := tokens.NewTokenClient(testAPIEndpoint.String())
		githubClient := github.NewClient(tokenClient)

		mux.HandleFunc("/users/blamewarrior", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"token": "test_token"}`))
		})

		mux.HandleFunc("/repos/blamewarrior/test/collaborators", func(w http.ResponseWriter, req *http.Request) {

			assert.Equal(t, "Bearer test_token", req.Header.Get("Authorization"))

			w.Write([]byte(`[{"login":"user1", "id": 1, "permissions": {"pull": true,"push": true,"admin": false}}]`))

		})

		collaboration := blamewarrior.NewCollaborationService()

		handler := main.NewFetchCollaboratorsHandler("blamewarrior.com", db, collaboration, githubClient)
		handler.GithubBaseURL = testAPIEndpoint

		handler.ServeHTTP(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))

		accounts, err := collaboration.ListAccounts(db, fmt.Sprintf("%s/%s", result.Owner, result.Name))
		require.NoError(t, err)

		require.Equal(t, len(result.Collaborators), len(accounts))

		for i := 0; i < len(result.Collaborators); i++ {
			assert.Equal(t, result.Collaborators[i].Login, accounts[i].Login)

			assert.Equal(t, result.Collaborators[i].Uid, accounts[i].Uid)
			assert.Equal(t, result.Collaborators[i].Permissions, accounts[i].Permissions)
		}

		teardownDB()
		teardownAPIServer()

	}
}

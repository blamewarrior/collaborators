/*
   Copyright (C) 2017 The BlameWarrior Authors.
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

package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blamewarrior/collaborators/blamewarrior"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/blamewarrior/collaborators"
)

func TestDisconnectCollaboratorHandler(t *testing.T) {
	results := []struct {
		Owner        string
		Name         string
		Collaborator string
		ResponseCode int
		ResponseBody string
	}{
		{
			Owner:        "",
			Name:         "",
			Collaborator: "test_collaborator",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: "Incorrect full name\n",
		},
		{
			Owner:        "blamewarrior",
			Name:         "test",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: "Incorrect collaborator name\n",
		},
		{
			Owner:        "blamewarrior",
			Name:         "test_disconnect_collaborator",
			Collaborator: "test_collaborator",
			ResponseCode: http.StatusNoContent,
			ResponseBody: "",
		},
	}

	for _, result := range results {
		db, teardown := setupTestDBConn()

		_, err := db.Exec("TRUNCATE repositories, collaboration, accounts")
		require.NoError(t, err)

		_, err = db.Exec(blamewarrior.CreateRepositoryQuery, fmt.Sprintf("%s/%s", result.Owner, result.Name))
		require.NoError(t, err)

		requestURL := fmt.Sprintf("/collaborators?:username=%s&:repo=%s&:collaborator=%s", result.Owner, result.Name, result.Collaborator)

		req, err := http.NewRequest("DELETE", requestURL, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		collaboration := blamewarrior.NewCollaborationService()

		handler := main.NewDisconnectCollaboratorHandler("blamewarrior.com", db, collaboration)
		handler.ServeHTTP(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))

		teardown()
	}
}

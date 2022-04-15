package intest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

func TestPerson(t *testing.T) {
	resourceName := "persons"
	startURL := appURL + "/" + resourceName

	var createdID string
	address := uuid.New().String()

	require.NoError(t, pgdao.New(db).JobsPurge(bgctx))
	require.NoError(t, pgdao.New(db).PersonsPurge(bgctx))

	t.Run("get•empty", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL, nil)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]model.Person, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("post", func(t *testing.T) {
		body := `{
			"address":"` + address + `"
		}`

		n := time.Now()

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Person)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			createdID = e.ID

			assert.True(t, strings.HasPrefix(res.Header.Get("location"), "/"+resourceName+"/"+e.ID))
			assert.NotEmpty(t, e.ID)
			assert.Equal(t, address, e.Address)
			// assert.NotEmpty(t, newObj.CreatedAt)
			assert.WithinDuration(t, n, e.CreatedAt, time.Since(n))

		}
	})

	t.Run("get", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL, nil)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]model.Person, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			if assert.NotEmpty(t, ee) {
				for _, e := range ee {
					assert.NotEmpty(t, e.ID)
					assert.NotEmpty(t, e.Address)
					assert.NotEmpty(t, e.CreatedAt)
				}
			}
		}
	})

	t.Run("get/:id", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL+"/"+createdID, nil)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Person)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, address, e.Address)
				assert.NotEmpty(t, e.CreatedAt)
			}
		}
	})

	t.Run("get/:id•404", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL+"/not-existent-entity", nil)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})
}

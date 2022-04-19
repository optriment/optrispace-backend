package intest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/web"
)

func TestJob(t *testing.T) {
	resourceName := "jobs"
	startURL := appURL + "/" + resourceName

	var createdID string

	require.NoError(t, pgdao.New(db).ApplicationsPurge(bgctx))
	require.NoError(t, pgdao.New(db).JobsPurge(bgctx))
	createdBy, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:      pgdao.NewID(),
		Address: pgdao.NewID() + pgdao.NewID(),
	})
	require.NoError(t, err)

	t.Run("get•empty", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL, nil)
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]model.Job, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("post•401", func(t *testing.T) {
		body := `{
			"title":"Create awesome site",
			"description": "There are words here. Very much words.",
			"budget": 100.2,
			"duration": 30
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Authorization required", e["message"])
		}
	})

	t.Run("post•full", func(t *testing.T) {
		body := `{
			"title":"Create awesome site",
			"description": "There are words here. Very much words.",
			"budget": "100.2",
			"duration": 30
		}`

		n := time.Now()

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			createdID = e.ID
			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.Equal(t, "Create awesome site", e.Title)
			assert.Equal(t, "There are words here. Very much words.", e.Description)
			assert.True(t, decimal.RequireFromString("100.2").Equal(e.Budget))
			assert.EqualValues(t, 30, e.Duration)
			assert.WithinDuration(t, n, e.CreatedAt, time.Since(n))
			assert.Equal(t, createdBy.ID, e.CreatedBy)
			assert.WithinDuration(t, n, e.UpdatedAt, time.Since(n))

			d, err := pgdao.New(db).JobGet(bgctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.Title, d.Title)
				assert.Equal(t, e.Description, d.Description)
				assert.Equal(t, "100.2", d.Budget.String)
				assert.Equal(t, e.Duration, d.Duration.Int32)

				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())

				assert.Equal(t, createdBy.ID, d.CreatedBy)
				assert.Equal(t, createdBy.Address, d.Address.String)
			}
		}
	})

	t.Run("post•wo-optional", func(t *testing.T) {
		body := `{
			"title":"Create awesome site (wo optional)",
			"description": "There are words here. Very much words. Without optional fields."
		}`

		n := time.Now()

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.Equal(t, "Create awesome site (wo optional)", e.Title)
			assert.Equal(t, "There are words here. Very much words. Without optional fields.", e.Description)
			assert.True(t, decimal.Zero.Equal(e.Budget))
			assert.EqualValues(t, 0, e.Duration)
			assert.WithinDuration(t, n, e.CreatedAt, time.Since(n))
			assert.Equal(t, createdBy.ID, e.CreatedBy)
			assert.WithinDuration(t, n, e.UpdatedAt, time.Since(n))

			d, err := pgdao.New(db).JobGet(bgctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.Title, d.Title)
				assert.Equal(t, e.Description, d.Description)
				assert.False(t, d.Budget.Valid)
				assert.False(t, d.Duration.Valid)

				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())

				assert.Equal(t, createdBy.ID, d.CreatedBy)
				assert.Equal(t, createdBy.Address, d.Address.String)
			}

		}
	})

	t.Run("post•required-fields", func(t *testing.T) {
		body := `{
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Regexp(t, "^Error while processing request:.*$", e["message"])
		}
	})

	t.Run("get", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL, nil)
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]model.Job, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			if assert.NotEmpty(t, ee) {
				for _, e := range ee {
					assert.NotEmpty(t, e.ID)
					assert.NotEmpty(t, e.Title)
					assert.NotEmpty(t, e.Description)
					assert.NotEmpty(t, e.CreatedAt)
					assert.NotEmpty(t, e.UpdatedAt)
					assert.NotEmpty(t, e.CreatedBy)
					assert.EqualValues(t, 0, e.ApplicationsCount)
				}
			}
		}
	})

	t.Run("get/:id", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL+"/"+createdID, nil)
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, "Create awesome site", e.Title)
				assert.Equal(t, "There are words here. Very much words.", e.Description)
				assert.True(t, decimal.RequireFromString("100.2").Equal(e.Budget))
				assert.EqualValues(t, 30, e.Duration)
				assert.NotEmpty(t, e.CreatedAt)
				assert.Equal(t, createdBy.ID, e.CreatedBy)
				assert.NotEmpty(t, e.UpdatedAt)
				assert.Equal(t, uint(0), e.ApplicationsCount)
			}
		}
	})

	t.Run("get/:id•not-found", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, startURL+"/"+"invalid-id", nil)
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Entity with specified id not found", e["message"])
		}
	})
}
